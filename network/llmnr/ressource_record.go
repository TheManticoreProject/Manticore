package llmnr

import (
	"encoding/binary"
	"fmt"
	"net"
)

// Resource Record Types
const (
	TypeA     uint16 = 1   // IPv4 address
	TypeNS    uint16 = 2   // Authoritative name server
	TypeCNAME uint16 = 5   // Canonical name for an alias
	TypeSOA   uint16 = 6   // Start of authority
	TypePTR   uint16 = 12  // Domain name pointer
	TypeMX    uint16 = 15  // Mail exchange
	TypeTXT   uint16 = 16  // Text strings
	TypeAAAA  uint16 = 28  // IPv6 address
	TypeSRV   uint16 = 33  // Service locator
	TypeOPT   uint16 = 41  // OPT pseudo-RR, RFC 2671
	TypeAXFR  uint16 = 252 // Transfer of entire zone
	TypeALL   uint16 = 255 // All records
)

func TypeToString(t uint16) string {
	switch t {
	case TypeA:
		return "A"
	case TypeNS:
		return "NS"
	case TypeCNAME:
		return "CNAME"
	case TypeSOA:
		return "SOA"
	case TypePTR:
		return "PTR"
	case TypeMX:
		return "MX"
	case TypeTXT:
		return "TXT"
	case TypeAAAA:
		return "AAAA"
	case TypeSRV:
		return "SRV"
	case TypeOPT:
		return "OPT"
	case TypeAXFR:
		return "AXFR"
	case TypeALL:
		return "ALL"
	}
	return "Unknown"
}

// Resource Record Classes
const (
	ClassIN      uint16 = 1      // Internet
	ClassCS      uint16 = 2      // CSNET (Obsolete)
	ClassCH      uint16 = 3      // CHAOS
	ClassHS      uint16 = 4      // Hesiod
	ClassNONE    uint16 = 254    // Used in dynamic update messages
	ClassANY     uint16 = 255    // Any class
	ClassQU      uint16 = 1      // QU flag for LLMNR (same as IN)
	ClassUNICAST uint16 = 0x8001 // LLMNR Unicast response
)

func ClassToString(c uint16) string {
	switch c {
	case ClassIN:
		return "IN"
	case ClassCS:
		return "CS"
	case ClassCH:
		return "CH"
	case ClassHS:
		return "HS"
	case ClassNONE:
		return "NONE"
	case ClassANY:
		return "ANY"
	// case ClassQU:
	// 	return "QU"
	case ClassUNICAST:
		return "UNICAST"
	}
	return "Unknown"
}

// ResourceRecord represents a resource record in the LLMNR protocol.
//
// A resource record is used to store information about a domain name, such as its IP address, mail server, or other
// related data. The ResourceRecord struct contains fields for the domain name, type, class, time-to-live (TTL),
// resource data length (RDLength), and resource data (RData).
//
// Fields:
// - Name: The domain name associated with the resource record.
// - Type: The type of the resource record, indicating the kind of data stored (e.g., TypeA for IPv4 address).
// - Class: The class of the resource record, typically ClassIN for Internet.
// - TTL: The time-to-live value, indicating how long the record can be cached before it should be discarded.
// - RDLength: The length of the resource data in bytes.
// - RData: The resource data, which contains the actual information associated with the domain name (e.g., an IP address).
//
// Usage example:
//
//	rr := ResourceRecord{
//	    Name:     "example.local",
//	    Type:     TypeA,
//	    Class:    ClassIN,
//	    TTL:      300,
//	    RDLength: 4,
//	    RData:    []byte{192, 168, 1, 1},
//	}
type ResourceRecord struct {
	Name     string `json:"name"`
	Type     uint16 `json:"type"`
	Class    uint16 `json:"class"`
	TTL      uint32 `json:"ttl"`
	RDLength uint16 `json:"rdlength"`
	RData    []byte `json:"rdata"`
}

// EncodeResourceRecord converts a ResourceRecord into a byte slice using the LLMNR protocol's wire format.
// It encodes the domain name, type, class, TTL, RDLength, and RData fields sequentially.
//
// Parameters:
// - rr: The ResourceRecord to be encoded.
//
// Returns:
//   - A byte slice with the encoded resource record.
//   - An error if encoding fails, such as with an invalid domain name.
//
// Example usage:
//
//	rr := ResourceRecord{
//	    Name:     "example.local",
//	    Type:     TypeA,
//	    Class:    ClassIN,
//	    TTL:      300,
//	    RDLength: 4,
//	    RData:    []byte{192, 168, 1, 1},
//	}
//	encodedBuf, err := EncodeResourceRecord(rr)
//	if err != nil {
//	    log.Fatalf("Failed to encode resource record: %v", err)
//	}
func EncodeResourceRecord(rr ResourceRecord) ([]byte, error) {
	buf := []byte{}

	nameBuf, err := EncodeDomainName(rr.Name)
	if err != nil {
		return nil, err
	}
	buf = append(buf, nameBuf...)

	bufferUint16 := make([]byte, 2)
	bufferUint32 := make([]byte, 4)

	binary.BigEndian.PutUint16(bufferUint16, rr.Type)
	buf = append(buf, bufferUint16...)

	binary.BigEndian.PutUint16(bufferUint16, rr.Class)
	buf = append(buf, bufferUint16...)

	binary.BigEndian.PutUint32(bufferUint32, rr.TTL)
	buf = append(buf, bufferUint32...)

	rr.RDLength = uint16(len(rr.RData))
	binary.BigEndian.PutUint16(bufferUint16, rr.RDLength)
	buf = append(buf, bufferUint16...)

	buf = append(buf, rr.RData...)

	return buf, nil
}

// DecodeResourceRecord decodes a byte slice into a ResourceRecord struct. It expects the byte slice to be in the wire format
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
//
// Usage example:
//
//	data := []byte{...} // byte slice containing the resource record in wire format
//	offset := 0
//	rr, newOffset, err := DecodeResourceRecord(data, offset)
//	if err != nil {
//	    log.Fatalf("Failed to decode resource record: %v", err)
//	}
//
// Start Generation Here
func DecodeResourceRecord(data []byte, offset int) (ResourceRecord, int, error) {
	var rr ResourceRecord
	var err error

	rr.Name, offset, err = DecodeDomainName(data, offset)
	if err != nil {
		return ResourceRecord{}, offset, err
	}

	if offset+10 > len(data) {
		return ResourceRecord{}, offset, fmt.Errorf("truncated resource record")
	}

	rr.Type = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	rr.Class = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	rr.TTL = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	rr.RDLength = binary.BigEndian.Uint16(data[offset:])
	offset += 2

	if offset+int(rr.RDLength) > len(data) {
		return ResourceRecord{}, offset, fmt.Errorf("truncated rdata")
	}

	rr.RData = make([]byte, rr.RDLength)
	copy(rr.RData, data[offset:offset+int(rr.RDLength)])
	offset += int(rr.RDLength)

	return rr, offset, nil
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
//
// Usage example:
//
//	rdata := IPToRData("192.168.1.1")
//	if rdata == nil {
//	    log.Fatalf("Invalid IP address")
//	}
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
//
// Usage example:
//
//	rdata := IPv4ToRData("192.168.1.1")
//	if rdata == nil {
//	    log.Fatalf("Invalid IPv4 address")
//	}
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
//
// Usage example:
//
//	rdata := IPv6ToRData("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
//	if rdata == nil {
//	    log.Fatalf("Invalid IPv6 address")
//	}
func IPv6ToRData(ip string) []byte {
	data := []byte{}

	addr := net.ParseIP(ip).To16()

	for _, b := range addr {
		data = append(data, byte(b))
	}

	return data
}
