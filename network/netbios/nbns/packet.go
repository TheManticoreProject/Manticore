package nbns

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// Label sequence encoding limits (RFC 1002 4.2.1.2 / RFC 1035 4.1.4).
const (
	// maxLabelLength is the largest label a single length octet can describe;
	// the top two bits are reserved as compression/escape flags.
	maxLabelLength = 0x3F // 63
	// labelPointerFlag marks a length octet whose top two bits are 0b11,
	// indicating a 14-bit message-relative compression pointer follows.
	labelPointerFlag = 0xC0
	// labelFlagMask isolates the two most-significant bits of a length octet.
	labelFlagMask = 0xC0
	// maxNamePointers bounds the number of compression pointers a single name
	// may follow, guarding against pointer loops and malicious chains.
	maxNamePointers = 128
)

// marshalName encodes a NetBIOS name as an RFC 1002 4.2.1.2 label sequence:
// the first label is the 32-byte first-level-encoded name, each scope
// component is a further length-prefixed label, and the sequence is
// terminated by a zero-length (0x00) root label. A name with no scope
// therefore encodes as a single 32-byte label followed by 0x00, matching
// the historical single-label wire form.
func marshalName(n *NetBIOSName) ([]byte, error) {
	encoded, err := n.FirstLevelEncode()
	if err != nil {
		return nil, err
	}

	var buf []byte
	for _, label := range strings.Split(encoded, ".") {
		if len(label) > maxLabelLength {
			return nil, fmt.Errorf("label too long: %d bytes (max %d)", len(label), maxLabelLength)
		}
		buf = append(buf, byte(len(label)))
		buf = append(buf, label...)
	}
	buf = append(buf, 0x00) // root label / terminator

	return buf, nil
}

// unmarshalName decodes an RFC 1002 4.2.1.2 label sequence starting at offset
// in data, following any RFC 1035 4.1.4 compression pointers (length octet
// with the top two bits set to 0b11, carrying a 14-bit message-relative
// offset). It returns the decoded name and the number of bytes consumed at
// the ORIGINAL offset: a name that begins with a pointer consumes exactly the
// two pointer octets regardless of how far the pointer chain reaches.
func unmarshalName(data []byte, offset int) (*NetBIOSName, int, error) {
	var labels []string

	pos := offset
	consumed := 0     // bytes consumed at the original call site
	followed := false // true once a pointer has been followed (stops advancing consumed)
	pointers := 0

	for {
		if pos >= len(data) {
			return nil, 0, fmt.Errorf("truncated name")
		}

		b := data[pos]

		if b&labelFlagMask == labelPointerFlag {
			// Compression pointer: a two-octet, 14-bit message-relative offset.
			if pos+1 >= len(data) {
				return nil, 0, fmt.Errorf("truncated compression pointer")
			}
			if !followed {
				consumed += 2
			}

			pointers++
			if pointers > maxNamePointers {
				return nil, 0, fmt.Errorf("too many compression pointers (possible loop)")
			}

			target := (int(b&0x3F) << 8) | int(data[pos+1])
			if target >= len(data) {
				return nil, 0, fmt.Errorf("compression pointer out of range: %d", target)
			}

			pos = target
			followed = true
			continue
		}

		if b&labelFlagMask != 0 {
			return nil, 0, fmt.Errorf("invalid label length flags: 0x%02x", b)
		}

		length := int(b)
		if !followed {
			consumed++
		}
		pos++

		if length == 0 {
			break // root label terminates the sequence
		}

		if pos+length > len(data) {
			return nil, 0, fmt.Errorf("truncated label")
		}

		labels = append(labels, string(data[pos:pos+length]))
		if !followed {
			consumed += length
		}
		pos += length
	}

	// Reassemble the first-level name and scope components ("name.scope...")
	// so FirstLevelDecode can split the encoded name from the scope ID.
	name, err := FirstLevelDecode(strings.Join(labels, "."))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode name: %v", err)
	}

	return name, consumed, nil
}

// Constants for packet types and flags
const (
	// Operation codes
	OpNameQuery    uint16 = 0x0000
	OpRegistration uint16 = 0x2800
	OpRelease      uint16 = 0x3000
	OpWACK         uint16 = 0x3800
	OpRefresh      uint16 = 0x4000
	OpRedirect     uint16 = 0x4800
	OpConflict     uint16 = 0x5000
	OpNodeStatus   uint16 = 0x2100

	// Response codes
	RcodeSuccess     uint16 = 0x0000
	RcodeFormatError uint16 = 0x0001
	RcodeServerError uint16 = 0x0002
	RcodeNameError   uint16 = 0x0003
	RcodeNotImpl     uint16 = 0x0004
	RcodeRefused     uint16 = 0x0005
	RcodeActive      uint16 = 0x0006
	RcodeConflict    uint16 = 0x0007

	// Header NM_FLAGS (RFC 1002 4.2.1.1: R | OPCODE | AA TC RD RA 0 0 B | RCODE)
	FlagResponse           uint16 = 0x8000 // R:  response
	FlagAuthoritative      uint16 = 0x0400 // AA: authoritative answer
	FlagTruncated          uint16 = 0x0200 // TC: truncated
	FlagRecursion          uint16 = 0x0100 // RD: recursion desired
	FlagRecursionAvailable uint16 = 0x0080 // RA: recursion available
	FlagBroadcast          uint16 = 0x0010 // B:  broadcast/multicast

	// RcodeMask isolates the RCODE field (low nibble) of the header flags.
	RcodeMask uint16 = 0x000F

	// Question Type
	QuestionTypeNB     uint16 = 0x0020
	QuestionTypeNBSTAT uint16 = 0x0021

	// Question Class
	QuestionClassIn uint16 = 0x0001 // Internet class
)

// NBNSHeader represents the header of a NetBIOS name service packet
type NBNSHeader struct {
	TransactionID uint16
	Flags         uint16
	Questions     uint16
	Answers       uint16
	Authority     uint16
	Additional    uint16
}

// NBNSQuestion represents a question section in a NetBIOS name service packet
type NBNSQuestion struct {
	Name  *NetBIOSName
	Type  uint16
	Class uint16
}

// NBNSResourceRecord represents a resource record in a NetBIOS name service packet
type NBNSResourceRecord struct {
	Name     *NetBIOSName
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []byte
}

// NBNSPacket represents a complete NetBIOS name service packet
type NBNSPacket struct {
	Header     NBNSHeader
	Questions  []NBNSQuestion
	Answers    []NBNSResourceRecord
	Authority  []NBNSResourceRecord
	Additional []NBNSResourceRecord
}

// Marshal encodes an NBNSPacket into a byte slice
func (p *NBNSPacket) Marshal() ([]byte, error) {
	buf := make([]byte, 12, 512) // Initial size for header, will grow as needed

	// Marshal header
	binary.BigEndian.PutUint16(buf[0:2], p.Header.TransactionID)
	binary.BigEndian.PutUint16(buf[2:4], p.Header.Flags)
	binary.BigEndian.PutUint16(buf[4:6], p.Header.Questions)
	binary.BigEndian.PutUint16(buf[6:8], p.Header.Answers)
	binary.BigEndian.PutUint16(buf[8:10], p.Header.Authority)
	binary.BigEndian.PutUint16(buf[10:12], p.Header.Additional)

	// Marshal questions
	for _, q := range p.Questions {
		encoded, err := marshalName(q.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to encode question name: %v", err)
		}

		// Add the name as an RFC 1002 4.2.1.2 label sequence
		buf = append(buf, encoded...)

		// Add type and class
		buf = binary.BigEndian.AppendUint16(buf, q.Type)
		buf = binary.BigEndian.AppendUint16(buf, q.Class)
	}

	// Marshal resource records (answers, authority, additional)
	for _, section := range [][]NBNSResourceRecord{p.Answers, p.Authority, p.Additional} {
		for _, rr := range section {
			encoded, err := marshalName(rr.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to encode resource record name: %v", err)
			}

			// Add the name as an RFC 1002 4.2.1.2 label sequence
			buf = append(buf, encoded...)

			// Add type, class, TTL, and RDATA length
			buf = binary.BigEndian.AppendUint16(buf, rr.Type)
			buf = binary.BigEndian.AppendUint16(buf, rr.Class)
			buf = binary.BigEndian.AppendUint32(buf, rr.TTL)
			buf = binary.BigEndian.AppendUint16(buf, rr.RDLength)

			// Add RDATA
			buf = append(buf, rr.RData...)
		}
	}

	return buf, nil
}

// Unmarshal decodes a byte slice into an NBNSPacket
func (p *NBNSPacket) Unmarshal(data []byte) (int, error) {
	if len(data) < 12 {
		return 0, fmt.Errorf("packet too short")
	}

	// Unmarshal header
	p.Header.TransactionID = binary.BigEndian.Uint16(data[0:2])
	p.Header.Flags = binary.BigEndian.Uint16(data[2:4])
	p.Header.Questions = binary.BigEndian.Uint16(data[4:6])
	p.Header.Answers = binary.BigEndian.Uint16(data[6:8])
	p.Header.Authority = binary.BigEndian.Uint16(data[8:10])
	p.Header.Additional = binary.BigEndian.Uint16(data[10:12])

	offset := 12

	// Unmarshal questions
	for i := uint16(0); i < p.Header.Questions; i++ {
		if offset >= len(data) {
			return 0, fmt.Errorf("truncated packet")
		}

		name, consumed, err := unmarshalName(data, offset)
		if err != nil {
			return 0, fmt.Errorf("failed to decode question name: %v", err)
		}
		offset += consumed

		if offset+4 > len(data) {
			return 0, fmt.Errorf("truncated question")
		}

		q := NBNSQuestion{
			Name:  name,
			Type:  binary.BigEndian.Uint16(data[offset : offset+2]),
			Class: binary.BigEndian.Uint16(data[offset+2 : offset+4]),
		}
		offset += 4

		p.Questions = append(p.Questions, q)
	}

	// Helper function to unmarshal resource records
	unmarshalRRs := func(count uint16) ([]NBNSResourceRecord, error) {
		var rrs []NBNSResourceRecord

		for i := uint16(0); i < count; i++ {
			if offset >= len(data) {
				return nil, fmt.Errorf("truncated packet")
			}

			name, consumed, err := unmarshalName(data, offset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode resource record name: %v", err)
			}
			offset += consumed

			if offset+10 > len(data) {
				return nil, fmt.Errorf("truncated resource record")
			}

			rr := NBNSResourceRecord{
				Name:     name,
				Type:     binary.BigEndian.Uint16(data[offset : offset+2]),
				Class:    binary.BigEndian.Uint16(data[offset+2 : offset+4]),
				TTL:      binary.BigEndian.Uint32(data[offset+4 : offset+8]),
				RDLength: binary.BigEndian.Uint16(data[offset+8 : offset+10]),
			}
			offset += 10

			if offset+int(rr.RDLength) > len(data) {
				return nil, fmt.Errorf("truncated RDATA")
			}

			rr.RData = make([]byte, rr.RDLength)
			copy(rr.RData, data[offset:offset+int(rr.RDLength)])
			offset += int(rr.RDLength)

			rrs = append(rrs, rr)
		}

		return rrs, nil
	}

	// Unmarshal answers, authority, and additional sections
	var err error

	p.Answers, err = unmarshalRRs(p.Header.Answers)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal answers: %v", err)
	}

	p.Authority, err = unmarshalRRs(p.Header.Authority)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal authority: %v", err)
	}

	p.Additional, err = unmarshalRRs(p.Header.Additional)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal additional: %v", err)
	}

	return len(data), nil
}
