package msdnsp

import (
	"encoding/binary"
	"fmt"
)

// DNSRecordVersion is the value that the Version field of a DNS_RECORD MUST carry.
const DNSRecordVersion uint8 = 0x05

// dnsRecordHeaderLength is the fixed size, in bytes, of the DNS_RECORD header that precedes
// the variable-length Data field: DataLength(2) + Type(2) + Version(1) + Rank(1) + Flags(2) +
// Serial(4) + TtlSeconds(4) + Reserved(4) + TimeStamp(4) = 24 bytes.
const dnsRecordHeaderLength = 24

// DNS_RECORD is the packed container that Active Directory stores in each value of the
// dnsRecord LDAP attribute. It holds a fixed 24-byte header followed by a variable-length Data
// field carrying the type-specific record-data payload (see DNS_RPC_RECORD_DATA,
// section 2.2.2.2.4) selected by the Type field.
//
// The multi-byte header fields are little-endian except TtlSeconds, which [MS-DNSP] mandates
// in big-endian byte order.
//
// Source: [MS-DNSP] dnsRecord (section 2.3.2.2)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/6912b338-5472-4f59-b912-0edb536b6ed8
type DNS_RECORD struct {
	// DataLength (2 bytes): An unsigned binary integer containing the length, in bytes, of the
	// Data field. Little-endian. Marshal recomputes this from the Data field.
	DataLength uint16

	// Type (2 bytes): The resource record's type. See DNS_RECORD_TYPE (section 2.2.2.1.1).
	// Little-endian.
	Type RecordType

	// Version (1 byte): The version number associated with the resource record attribute. The
	// value MUST be 0x05.
	Version uint8

	// Rank (1 byte): The least-significant byte of one of the RANK* flag values (see dwFlags,
	// section 2.2.2.2.5).
	Rank uint8

	// Flags (2 bytes): Not used. The value MUST be 0x0000. Little-endian.
	Flags uint16

	// Serial (4 bytes): The serial number of the SOA record of the zone containing this
	// resource record. Little-endian.
	Serial uint32

	// TtlSeconds (4 bytes): The record's time-to-live, in seconds. This field uses big-endian
	// byte order.
	TtlSeconds uint32

	// Reserved (4 bytes): This field is reserved for future use. The value MUST be 0x00000000.
	Reserved uint32

	// TimeStamp (4 bytes): The time at which the record was last refreshed, in hours since
	// 1601-01-01 00:00:00 UTC, or 0 for a static record. Little-endian.
	TimeStamp uint32

	// Data (variable): The resource record's data. See DNS_RPC_RECORD_DATA (section 2.2.2.2.4).
	Data []byte
}

// NewDNS_RECORD creates a new DNS_RECORD with the Version field initialized to the mandatory
// value 0x05.
//
// Returns:
// - A pointer to the new DNS_RECORD structure
func NewDNS_RECORD() *DNS_RECORD {
	return &DNS_RECORD{
		Version: DNSRecordVersion,
	}
}

// SetData replaces the record's Data field and updates DataLength to match. payload is any of
// the record-data payload structures (for example *DNS_RPC_RECORD_A); its Marshal output
// becomes the Data field. The caller is responsible for setting Type to match the payload.
//
// Parameters:
// - payload: A record-data payload exposing Marshal() ([]byte, error)
//
// Returns:
// - An error if the payload fails to marshal
func (r *DNS_RECORD) SetData(payload interface{ Marshal() ([]byte, error) }) error {
	data, err := payload.Marshal()
	if err != nil {
		return fmt.Errorf("marshalling record data: %w", err)
	}
	if len(data) > 0xFFFF {
		return fmt.Errorf("record data is %d bytes, exceeds the 65535-byte DataLength limit", len(data))
	}
	r.Data = data
	r.DataLength = uint16(len(data))
	return nil
}

// Marshal marshals the DNS_RECORD structure into a byte array. DataLength is recomputed from
// the current length of the Data field so the header always matches the payload.
//
// Returns:
// - A byte array representing the DNS_RECORD structure
// - An error if the marshaling fails
func (r *DNS_RECORD) Marshal() ([]byte, error) {
	if len(r.Data) > 0xFFFF {
		return nil, fmt.Errorf("record data is %d bytes, exceeds the 65535-byte DataLength limit", len(r.Data))
	}
	r.DataLength = uint16(len(r.Data))

	marshalled := make([]byte, dnsRecordHeaderLength)
	binary.LittleEndian.PutUint16(marshalled[0:2], r.DataLength)
	binary.LittleEndian.PutUint16(marshalled[2:4], uint16(r.Type))
	marshalled[4] = r.Version
	marshalled[5] = r.Rank
	binary.LittleEndian.PutUint16(marshalled[6:8], r.Flags)
	binary.LittleEndian.PutUint32(marshalled[8:12], r.Serial)
	// TtlSeconds is big-endian per [MS-DNSP] section 2.3.2.2.
	binary.BigEndian.PutUint32(marshalled[12:16], r.TtlSeconds)
	binary.LittleEndian.PutUint32(marshalled[16:20], r.Reserved)
	binary.LittleEndian.PutUint32(marshalled[20:24], r.TimeStamp)

	marshalled = append(marshalled, r.Data...)
	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_RECORD structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read (24-byte header plus DataLength bytes of Data)
// - An error if the unmarshaling fails
func (r *DNS_RECORD) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < dnsRecordHeaderLength {
		return 0, fmt.Errorf("rawData too short for DNS_RECORD header: need %d bytes, have %d", dnsRecordHeaderLength, len(rawData))
	}

	r.DataLength = binary.LittleEndian.Uint16(rawData[0:2])
	r.Type = RecordType(binary.LittleEndian.Uint16(rawData[2:4]))
	r.Version = rawData[4]
	r.Rank = rawData[5]
	r.Flags = binary.LittleEndian.Uint16(rawData[6:8])
	r.Serial = binary.LittleEndian.Uint32(rawData[8:12])
	// TtlSeconds is big-endian per [MS-DNSP] section 2.3.2.2.
	r.TtlSeconds = binary.BigEndian.Uint32(rawData[12:16])
	r.Reserved = binary.LittleEndian.Uint32(rawData[16:20])
	r.TimeStamp = binary.LittleEndian.Uint32(rawData[20:24])
	offset := dnsRecordHeaderLength

	if len(rawData) < offset+int(r.DataLength) {
		return 0, fmt.Errorf("rawData too short for DNS_RECORD Data: need %d bytes, have %d", offset+int(r.DataLength), len(rawData))
	}

	if r.DataLength == 0 {
		r.Data = nil
	} else {
		r.Data = make([]byte, r.DataLength)
		copy(r.Data, rawData[offset:offset+int(r.DataLength)])
	}
	offset += int(r.DataLength)

	return offset, nil
}
