package resourcerecord

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/errors"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
)

// ResourceRecord represents a resource record in the LLMNR protocol.
//
// A resource record is used to store information about a domain name, such as its IP address, mail server, or other
// related data. The ResourceRecord struct contains fields for the domain name, type, class, time-to-live (TTL),
// resource data length (RDLength), and resource data (RData).
type ResourceRecord struct {
	// The domain name associated with the resource record.
	Name domain_name.DomainName `json:"name"`

	// The type of the resource record, indicating the kind of data stored (e.g., TypeA for IPv4 address).
	Type llmnr_type.Type `json:"type"`

	// The class of the resource record, typically ClassIN for Internet.
	Class class.Class `json:"class"`

	// The time-to-live value, indicating how long the record can be cached before it should be discarded.
	TTL uint32 `json:"ttl"`

	// The length of the resource data in bytes.
	RDLength uint16 `json:"rdlength"`

	// The resource data, which contains the actual information associated with the domain name (e.g., an IP address).
	RData []byte `json:"rdata"`

	// message and rdataOffset retain the context needed to resolve 0xC0
	// compression pointers embedded inside name-bearing RDATA (PTR/CNAME/NS/SRV).
	// They are populated by UnmarshalFromMessage: message is the full LLMNR
	// message buffer and rdataOffset is the absolute offset of RData within it.
	// When a record is built by hand (rather than decoded from a message) these
	// stay zero-valued and the typed accessors fall back to decoding names from
	// RData alone, which resolves correctly only for uncompressed names.
	message     []byte
	rdataOffset int
}

// Marshal converts a ResourceRecord into a byte slice using the LLMNR protocol's wire format.
// It encodes the domain name, type, class, TTL, RDLength, and RData fields sequentially.
//
// Parameters:
// - rr: The ResourceRecord to be encoded.
//
// Returns:
//   - A byte slice with the encoded resource record.
//   - An error if encoding fails, such as with an invalid domain name.
func (rr *ResourceRecord) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	nameBuf, err := rr.Name.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, nameBuf...)

	bufferUint16 := make([]byte, 2)
	bufferUint32 := make([]byte, 4)

	binary.BigEndian.PutUint16(bufferUint16, uint16(rr.Type))
	marshalledData = append(marshalledData, bufferUint16...)

	binary.BigEndian.PutUint16(bufferUint16, uint16(rr.Class))
	marshalledData = append(marshalledData, bufferUint16...)

	binary.BigEndian.PutUint32(bufferUint32, rr.TTL)
	marshalledData = append(marshalledData, bufferUint32...)

	rr.RDLength = uint16(len(rr.RData))
	binary.BigEndian.PutUint16(bufferUint16, rr.RDLength)
	marshalledData = append(marshalledData, bufferUint16...)

	marshalledData = append(marshalledData, rr.RData...)

	return marshalledData, nil
}

// Unmarshal decodes a byte slice into a ResourceRecord struct. It expects the byte slice to be in the wire format
// as specified by the LLMNR protocol. The function first decodes the domain name, followed by the type, class, TTL, RDLength, and RData fields.
//
// Parameters:
// - data: A byte slice containing the resource record in wire format.
// - offset: The starting position in the byte slice from which to begin decoding.
//
// Returns:
//   - A ResourceRecord struct containing the decoded data.
//   - An integer representing the new offset position after decoding.
//   - An error if the decoding fails at any point, such as if the data is too short or if there is an error decoding the domain name.
func (rr *ResourceRecord) Unmarshal(data []byte) (int, error) {
	return rr.UnmarshalFromMessage(data, 0)
}

// UnmarshalFromMessage decodes a ResourceRecord located at offset inside the
// full message buffer. It is used instead of Unmarshal when the record's name
// may contain 0xC0 compression pointers (responders routinely compress the
// answer NAME to point back at the question), so those pointers resolve against
// the start of the message (RFC 1035 §4.1.4) rather than against a sub-slice.
//
// Parameters:
// - message: The full LLMNR message in wire format.
// - offset: The absolute offset of the resource record inside message.
//
// Returns:
//   - The number of bytes consumed from message at offset.
//   - An error if the decoding fails at any point.
func (rr *ResourceRecord) UnmarshalFromMessage(message []byte, offset int) (int, error) {
	start := offset

	bytesReadName, err := rr.Name.UnmarshalFromMessage(message, offset)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling name: %w: %w", err, errors.ErrInvalidResourceRecord)
	}
	offset += bytesReadName

	if offset+10 > len(message) {
		return 0, fmt.Errorf("truncated resource record: %w", errors.ErrInvalidResourceRecord)
	}

	rr.Type = llmnr_type.Type(binary.BigEndian.Uint16(message[offset:]))
	offset += 2
	rr.Class = class.Class(binary.BigEndian.Uint16(message[offset:]))
	offset += 2
	rr.TTL = binary.BigEndian.Uint32(message[offset:])
	offset += 4
	rr.RDLength = binary.BigEndian.Uint16(message[offset:])
	offset += 2

	if offset+int(rr.RDLength) > len(message) {
		return 0, fmt.Errorf("truncated rdata: %w", errors.ErrInvalidResourceRecord)
	}

	// Retain the message context so that name-bearing RDATA (PTR/CNAME/NS/SRV)
	// can resolve compression pointers relative to the start of the message
	// (RFC 1035 §4.1.4) via the typed accessors below.
	rr.message = message
	rr.rdataOffset = offset

	rr.RData = make([]byte, rr.RDLength)
	copy(rr.RData, message[offset:offset+int(rr.RDLength)])
	offset += int(rr.RDLength)

	return offset - start, nil
}

// IPToRData converts an IP address string to its corresponding RData byte slice representation.
// It determines whether the IP address is IPv4 or IPv6 and calls the appropriate conversion function.
//
// Parameters:
// - ip: A string representing the IP address to be converted.
//
// Returns:
// - A byte slice containing the RData representation of the IP address.
// - nil if the IP address is neither a valid IPv4 nor IPv6 address.
func IPToRData(ip string) []byte {
	if net.ParseIP(ip).To4() != nil {
		return IPv4ToRData(ip)
	}

	if net.ParseIP(ip).To16() != nil {
		return IPv6ToRData(ip)
	}

	return nil
}

// IPv4ToRData converts an IPv4 address string to its corresponding RData byte slice representation.
//
// Parameters:
// - ip: A string representing the IPv4 address to be converted.
//
// Returns:
// - A byte slice containing the RData representation of the IPv4 address.
func IPv4ToRData(ip string) []byte {
	data := []byte{}

	addr := net.ParseIP(ip).To4()

	for _, b := range addr {
		data = append(data, byte(b))
	}

	return data
}

// IPv6ToRData converts an IPv6 address string to its corresponding RData byte slice representation.
//
// Parameters:
// - ip: A string representing the IPv6 address to be converted.
//
// Returns:
// - A byte slice containing the RData representation of the IPv6 address.
func IPv6ToRData(ip string) []byte {
	data := []byte{}

	addr := net.ParseIP(ip).To16()

	for _, b := range addr {
		data = append(data, byte(b))
	}

	return data
}

// Describe prints a detailed description of the ResourceRecord.
//
// Parameters:
// - indent: An integer value specifying the indentation level for the output.
func (rr *ResourceRecord) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<ResourceRecord>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mName\x1b[0m: %s\n", indentPrompt, rr.Name)
	fmt.Printf("%s │ \x1b[93mType\x1b[0m: %s (0x%04x)\n", indentPrompt, rr.Type.String(), rr.Type)
	fmt.Printf("%s │ \x1b[93mClass\x1b[0m: %s (0x%04x)\n", indentPrompt, rr.Class.String(), rr.Class)
	fmt.Printf("%s │ \x1b[93mTTL\x1b[0m: %d\n", indentPrompt, rr.TTL)
	fmt.Printf("%s │ \x1b[93mRDLength\x1b[0m: %d\n", indentPrompt, rr.RDLength)
	if len(rr.RData) > 0 {
		// Try to present human-friendly data for common record types
		switch rr.Type {
		case llmnr_type.TypeA, llmnr_type.TypeAAAA:
			ip := net.IP(rr.RData)
			if (rr.Type == llmnr_type.TypeA && len(rr.RData) == net.IPv4len) || (rr.Type == llmnr_type.TypeAAAA && len(rr.RData) == net.IPv6len) {
				fmt.Printf("%s │ \x1b[93mRData\x1b[0m: %s\n", indentPrompt, ip.String())
				break
			}
			fallthrough
		default:
			fmt.Printf("%s │ \x1b[93mRData\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(rr.RData))
		}
	}
	fmt.Printf("%s └───\n", indentPrompt)
}
