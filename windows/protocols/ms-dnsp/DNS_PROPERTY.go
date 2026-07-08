package msdnsp

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// dnsPropertyHeaderLength is the fixed size, in bytes, of the DNS_PROPERTY header that precedes
// the variable-length Data field: DataLength(4) + NameLength(4) + Flag(4) + Version(4) + Id(4)
// = 20 bytes.
const dnsPropertyHeaderLength = 20

// DNSPropertyVersion is the value that the Version field of a DNS_PROPERTY MUST carry.
const DNSPropertyVersion uint32 = 0x00000001

// DNS_PROPERTY is the packed container that Active Directory stores in each value of the
// dnsProperty LDAP attribute of a DNS zone object. It holds a fixed 20-byte header, a
// variable-length Data field of DataLength bytes carrying the property value selected by the Id
// field (see PropertyId, section 2.3.2.1.1), and a trailing 1-byte Name field that is not used.
//
// All header fields are little-endian.
//
// Source: [MS-DNSP] dnsProperty (section 2.3.2.1)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/445c7843-e4a1-4222-8c0f-630c230a4c80
type DNS_PROPERTY struct {
	// DataLength (4 bytes): The length, in bytes, of the Data field. Marshal recomputes this
	// from the Data field. If 0, the property carries its default value.
	DataLength uint32

	// NameLength (4 bytes): Not used. The value MUST be ignored and assumed to be 0x00000001.
	NameLength uint32

	// Flag (4 bytes): Reserved for future use. The value MUST be 0x00000000.
	Flag uint32

	// Version (4 bytes): The version of the property attribute. The value MUST be 0x00000001.
	Version uint32

	// Id (4 bytes): The property's type, selecting how Data is interpreted.
	Id PropertyId

	// Data (variable): The property value. See the As* accessors for interpretation.
	Data []byte

	// Name (1 byte): Not used. The value MUST be of length 1 byte and MUST be ignored.
	Name uint8
}

// NewDNS_PROPERTY creates a new DNS_PROPERTY with Version and NameLength initialized to the
// values mandated by [MS-DNSP] section 2.3.2.1 (0x00000001 each).
//
// Returns:
// - A pointer to the new DNS_PROPERTY structure
func NewDNS_PROPERTY() *DNS_PROPERTY {
	return &DNS_PROPERTY{
		NameLength: 0x00000001,
		Version:    DNSPropertyVersion,
	}
}

// Marshal marshals the DNS_PROPERTY structure into a byte array. DataLength is recomputed from
// the current length of the Data field so the header always matches the payload.
//
// Returns:
// - A byte array representing the DNS_PROPERTY structure
// - An error if the marshaling fails
func (p *DNS_PROPERTY) Marshal() ([]byte, error) {
	// Compare in uint64 so the 0xFFFFFFFF bound does not overflow int on 32-bit platforms
	// (where int is 32 bits and len returns int).
	if uint64(len(p.Data)) > 0xFFFFFFFF {
		return nil, fmt.Errorf("property data is %d bytes, exceeds the 4294967295-byte DataLength limit", len(p.Data))
	}
	p.DataLength = uint32(len(p.Data))

	marshalled := make([]byte, dnsPropertyHeaderLength)
	binary.LittleEndian.PutUint32(marshalled[0:4], p.DataLength)
	binary.LittleEndian.PutUint32(marshalled[4:8], p.NameLength)
	binary.LittleEndian.PutUint32(marshalled[8:12], p.Flag)
	binary.LittleEndian.PutUint32(marshalled[12:16], p.Version)
	binary.LittleEndian.PutUint32(marshalled[16:20], uint32(p.Id))

	marshalled = append(marshalled, p.Data...)
	// The trailing Name field is a single, unused byte.
	marshalled = append(marshalled, p.Name)
	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_PROPERTY structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read (20-byte header, DataLength bytes of Data, and the 1-byte Name)
// - An error if the unmarshaling fails
func (p *DNS_PROPERTY) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < dnsPropertyHeaderLength {
		return 0, fmt.Errorf("rawData too short for DNS_PROPERTY header: need %d bytes, have %d", dnsPropertyHeaderLength, len(rawData))
	}

	p.DataLength = binary.LittleEndian.Uint32(rawData[0:4])
	p.NameLength = binary.LittleEndian.Uint32(rawData[4:8])
	p.Flag = binary.LittleEndian.Uint32(rawData[8:12])
	p.Version = binary.LittleEndian.Uint32(rawData[12:16])
	p.Id = PropertyId(binary.LittleEndian.Uint32(rawData[16:20]))
	offset := dnsPropertyHeaderLength

	// The header, the Data field, and the trailing 1-byte Name field must all be present.
	if len(rawData) < offset+int(p.DataLength)+1 {
		return 0, fmt.Errorf("rawData too short for DNS_PROPERTY Data and Name: need %d bytes, have %d", offset+int(p.DataLength)+1, len(rawData))
	}

	if p.DataLength == 0 {
		p.Data = nil
	} else {
		p.Data = make([]byte, p.DataLength)
		copy(p.Data, rawData[offset:offset+int(p.DataLength)])
	}
	offset += int(p.DataLength)

	p.Name = rawData[offset]
	offset++

	return offset, nil
}

// AsUint32 interprets the Data field as a little-endian integer, the form used by the scalar
// zone properties (DSPROPERTY_ZONE_TYPE, DSPROPERTY_ZONE_ALLOW_UPDATE, the interval properties,
// DSPROPERTY_ZONE_AGING_STATE, DSPROPERTY_ZONE_DCPROMO_CONVERT, and DSPROPERTY_ZONE_NODE_DBFLAGS).
//
// Although [MS-DNSP] types these properties as a DWORD, a Windows DNS server stores them in the
// minimum number of little-endian bytes: for example DSPROPERTY_ZONE_ALLOW_UPDATE is commonly a
// single byte. AsUint32 therefore reads up to four bytes and zero-extends, rather than requiring
// exactly four.
//
// Returns:
// - The value, read from up to the first four bytes of Data (little-endian, zero-extended)
// - An error if Data is empty
func (p *DNS_PROPERTY) AsUint32() (uint32, error) {
	if len(p.Data) == 0 {
		return 0, fmt.Errorf("DNS_PROPERTY %s: Data is empty, need at least 1 byte for an integer", p.Id)
	}
	n := len(p.Data)
	if n > 4 {
		n = 4
	}
	var v uint32
	for i := 0; i < n; i++ {
		v |= uint32(p.Data[i]) << (8 * i)
	}
	return v, nil
}

// AsZoneType interprets the Data field as a ZoneType (dwZoneType). It is meaningful for a
// DSPROPERTY_ZONE_TYPE property.
//
// Returns:
// - The zone type
// - An error if Data is empty
func (p *DNS_PROPERTY) AsZoneType() (ZoneType, error) {
	v, err := p.AsUint32()
	if err != nil {
		return 0, err
	}
	return ZoneType(v), nil
}

// AsZoneUpdate interprets the Data field as a ZoneUpdate (fAllowUpdate). It is meaningful for a
// DSPROPERTY_ZONE_ALLOW_UPDATE property.
//
// Returns:
// - The dynamic-update policy
// - An error if Data is empty
func (p *DNS_PROPERTY) AsZoneUpdate() (ZoneUpdate, error) {
	v, err := p.AsUint32()
	if err != nil {
		return 0, err
	}
	return ZoneUpdate(v), nil
}

// AsIP4Array interprets the Data field as an IP4_ARRAY (section 2.2.3.2.1): a little-endian
// AddrCount followed by AddrCount IPv4 addresses, each a 4-byte value in network byte order. It
// is meaningful for the DSPROPERTY_ZONE_SCAVENGING_SERVERS, DSPROPERTY_ZONE_MASTER_SERVERS, and
// DSPROPERTY_ZONE_AUTO_NS_SERVERS properties.
//
// Returns:
// - The IPv4 addresses (an empty, non-nil slice when AddrCount is 0)
// - An error if Data is malformed
func (p *DNS_PROPERTY) AsIP4Array() ([]net.IP, error) {
	if len(p.Data) < 4 {
		return nil, fmt.Errorf("DNS_PROPERTY %s: Data is %d bytes, need at least 4 for an IP4_ARRAY", p.Id, len(p.Data))
	}
	count := binary.LittleEndian.Uint32(p.Data[0:4])
	if uint64(len(p.Data)) < 4+uint64(count)*4 {
		return nil, fmt.Errorf("DNS_PROPERTY %s: IP4_ARRAY declares %d addresses but Data holds only %d bytes", p.Id, count, len(p.Data))
	}
	addrs := make([]net.IP, 0, count)
	offset := 4
	for i := uint32(0); i < count; i++ {
		b := p.Data[offset : offset+4]
		addrs = append(addrs, net.IPv4(b[0], b[1], b[2], b[3]))
		offset += 4
	}
	return addrs, nil
}

// AsUTF16String interprets the Data field as a null-terminated little-endian UTF-16 string, the
// form used by the DSPROPERTY_ZONE_DELETED_FROM_HOSTNAME property. The terminating NUL, if
// present, is removed.
//
// Returns:
// - The decoded string
func (p *DNS_PROPERTY) AsUTF16String() string {
	return strings.TrimRight(utf16.DecodeUTF16LE(p.Data), "\x00")
}

// String returns a human-readable, one-line rendering of the property: its Id name followed by
// an interpretation of the Data field appropriate to that Id, falling back to the raw byte
// count for property types whose payload is not decoded here.
func (p *DNS_PROPERTY) String() string {
	switch p.Id {
	case DSPROPERTY_ZONE_TYPE:
		if v, err := p.AsZoneType(); err == nil {
			return fmt.Sprintf("%s = %s", p.Id, v)
		}
	case DSPROPERTY_ZONE_ALLOW_UPDATE:
		if v, err := p.AsZoneUpdate(); err == nil {
			return fmt.Sprintf("%s = %s", p.Id, v)
		}
	case DSPROPERTY_ZONE_NOREFRESH_INTERVAL, DSPROPERTY_ZONE_REFRESH_INTERVAL,
		DSPROPERTY_ZONE_AGING_STATE, DSPROPERTY_ZONE_AGING_ENABLED_TIME,
		DSPROPERTY_ZONE_DCPROMO_CONVERT, DSPROPERTY_ZONE_NODE_DBFLAGS:
		if v, err := p.AsUint32(); err == nil {
			return fmt.Sprintf("%s = %d", p.Id, v)
		}
	case DSPROPERTY_ZONE_SCAVENGING_SERVERS, DSPROPERTY_ZONE_MASTER_SERVERS,
		DSPROPERTY_ZONE_AUTO_NS_SERVERS:
		if addrs, err := p.AsIP4Array(); err == nil {
			ss := make([]string, len(addrs))
			for i, a := range addrs {
				ss[i] = a.String()
			}
			return fmt.Sprintf("%s = [%s]", p.Id, strings.Join(ss, ", "))
		}
	case DSPROPERTY_ZONE_DELETED_FROM_HOSTNAME:
		return fmt.Sprintf("%s = %q", p.Id, p.AsUTF16String())
	}
	return fmt.Sprintf("%s = <%d bytes>", p.Id, len(p.Data))
}
