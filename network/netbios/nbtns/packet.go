package nbtns

import (
	"encoding/binary"
	"fmt"
)

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

	// Flags
	FlagResponse      uint16 = 0x8000
	FlagAuthoritative uint16 = 0x0400
	FlagTruncated     uint16 = 0x0200
	FlagRecursion     uint16 = 0x0100
	FlagBroadcast     uint16 = 0x0010
)

// NBTNSHeader represents the header of a NetBIOS name service packet
type NBTNSHeader struct {
	TransactionID uint16
	Flags         uint16
	Questions     uint16
	Answers       uint16
	Authority     uint16
	Additional    uint16
}

// NBTNSQuestion represents a question section in a NetBIOS name service packet
type NBTNSQuestion struct {
	Name  *NetBIOSName
	Type  uint16
	Class uint16
}

// NBTNSResourceRecord represents a resource record in a NetBIOS name service packet
type NBTNSResourceRecord struct {
	Name     *NetBIOSName
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []byte
}

// NBTNSPacket represents a complete NetBIOS name service packet
type NBTNSPacket struct {
	Header     NBTNSHeader
	Questions  []NBTNSQuestion
	Answers    []NBTNSResourceRecord
	Authority  []NBTNSResourceRecord
	Additional []NBTNSResourceRecord
}

// Marshal encodes an NBTNSPacket into a byte slice
func (p *NBTNSPacket) Marshal() ([]byte, error) {
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
		encoded, err := q.Name.FirstLevelEncode()
		if err != nil {
			return nil, fmt.Errorf("failed to encode question name: %v", err)
		}

		// Add name length and name
		buf = append(buf, byte(len(encoded)))
		buf = append(buf, []byte(encoded)...)

		// Add type and class
		buf = binary.BigEndian.AppendUint16(buf, q.Type)
		buf = binary.BigEndian.AppendUint16(buf, q.Class)
	}

	// Marshal resource records (answers, authority, additional)
	for _, section := range [][]NBTNSResourceRecord{p.Answers, p.Authority, p.Additional} {
		for _, rr := range section {
			encoded, err := rr.Name.FirstLevelEncode()
			if err != nil {
				return nil, fmt.Errorf("failed to encode resource record name: %v", err)
			}

			// Add name length and name
			buf = append(buf, byte(len(encoded)))
			buf = append(buf, []byte(encoded)...)

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

// Unmarshal decodes a byte slice into an NBTNSPacket
func (p *NBTNSPacket) Unmarshal(data []byte) (int, error) {
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

		nameLen := int(data[offset])
		offset++

		if offset+nameLen > len(data) {
			return 0, fmt.Errorf("truncated name")
		}

		name, err := FirstLevelDecode(string(data[offset : offset+nameLen]))
		if err != nil {
			return 0, fmt.Errorf("failed to decode name: %v", err)
		}
		offset += nameLen

		if offset+4 > len(data) {
			return 0, fmt.Errorf("truncated question")
		}

		q := NBTNSQuestion{
			Name:  name,
			Type:  binary.BigEndian.Uint16(data[offset : offset+2]),
			Class: binary.BigEndian.Uint16(data[offset+2 : offset+4]),
		}
		offset += 4

		p.Questions = append(p.Questions, q)
	}

	// Helper function to unmarshal resource records
	unmarshalRRs := func(count uint16) ([]NBTNSResourceRecord, error) {
		var rrs []NBTNSResourceRecord

		for i := uint16(0); i < count; i++ {
			if offset >= len(data) {
				return nil, fmt.Errorf("truncated packet")
			}

			nameLen := int(data[offset])
			offset++

			if offset+nameLen > len(data) {
				return nil, fmt.Errorf("truncated name")
			}

			name, err := FirstLevelDecode(string(data[offset : offset+nameLen]))
			if err != nil {
				return nil, fmt.Errorf("failed to decode name: %v", err)
			}
			offset += nameLen

			if offset+10 > len(data) {
				return nil, fmt.Errorf("truncated resource record")
			}

			rr := NBTNSResourceRecord{
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
