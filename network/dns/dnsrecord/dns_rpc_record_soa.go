package dnsrecord

import (
	"encoding/binary"
	"fmt"
)

// DNS_RPC_RECORD_SOA is the record-data payload for a DNS_TYPE_SOA (Start of Authority)
// record, as specified in section 3.3.13 of [RFC1035].
//
// The five numeric fields are stored in network byte order (big-endian). The two names are
// stored in DNS_COUNT_NAME (section 2.2.2.2.2) form in a dnsRecord attribute value carried
// over LDAP (the specification documents them in DNS_RPC_NAME form for the RPC wire).
//
// Source: [MS-DNSP] DNS_RPC_RECORD_SOA (section 2.2.2.2.4.3)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/dcd3ec16-d6bf-4bb4-9128-6172f9e5f066
type DNS_RPC_RECORD_SOA struct {
	// DwSerialNo (4 bytes): The serial number of the SOA record. Big-endian.
	DwSerialNo uint32

	// DwRefresh (4 bytes): The interval, in seconds, at which a secondary DNS server attempts
	// to contact the primary DNS server for an update. Big-endian.
	DwRefresh uint32

	// DwRetry (4 bytes): The interval, in seconds, at which a secondary DNS server retries to
	// check with the primary DNS server in case of failure. Big-endian.
	DwRetry uint32

	// DwExpire (4 bytes): The duration, in seconds, that a secondary DNS server continues to
	// attempt updates before assuming the primary DNS server is unreachable. Big-endian.
	DwExpire uint32

	// DwMinimumTtl (4 bytes): The minimum duration, in seconds, for which record data in the
	// zone is valid. Big-endian.
	DwMinimumTtl uint32

	// NamePrimaryServer (variable): The FQDN of the primary DNS server for this zone.
	NamePrimaryServer DNS_COUNT_NAME

	// ZoneAdminEmail (variable): The contact email address for the zone administrators.
	ZoneAdminEmail DNS_COUNT_NAME
}

// NewDNS_RPC_RECORD_SOA creates a new, empty DNS_RPC_RECORD_SOA.
//
// Returns:
// - A pointer to the new DNS_RPC_RECORD_SOA structure
func NewDNS_RPC_RECORD_SOA() *DNS_RPC_RECORD_SOA {
	return &DNS_RPC_RECORD_SOA{}
}

// Marshal marshals the DNS_RPC_RECORD_SOA structure into a byte array.
//
// Returns:
// - A byte array representing the DNS_RPC_RECORD_SOA structure
// - An error if the marshaling fails
func (r *DNS_RPC_RECORD_SOA) Marshal() ([]byte, error) {
	marshalled := make([]byte, 20)
	binary.BigEndian.PutUint32(marshalled[0:4], r.DwSerialNo)
	binary.BigEndian.PutUint32(marshalled[4:8], r.DwRefresh)
	binary.BigEndian.PutUint32(marshalled[8:12], r.DwRetry)
	binary.BigEndian.PutUint32(marshalled[12:16], r.DwExpire)
	binary.BigEndian.PutUint32(marshalled[16:20], r.DwMinimumTtl)

	primary, err := r.NamePrimaryServer.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshalling NamePrimaryServer: %w", err)
	}
	marshalled = append(marshalled, primary...)

	admin, err := r.ZoneAdminEmail.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshalling ZoneAdminEmail: %w", err)
	}
	marshalled = append(marshalled, admin...)

	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_RPC_RECORD_SOA structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read
// - An error if the unmarshaling fails
func (r *DNS_RPC_RECORD_SOA) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < 20 {
		return 0, fmt.Errorf("rawData too short for DNS_RPC_RECORD_SOA fixed fields: need 20 bytes, have %d", len(rawData))
	}

	r.DwSerialNo = binary.BigEndian.Uint32(rawData[0:4])
	r.DwRefresh = binary.BigEndian.Uint32(rawData[4:8])
	r.DwRetry = binary.BigEndian.Uint32(rawData[8:12])
	r.DwExpire = binary.BigEndian.Uint32(rawData[12:16])
	r.DwMinimumTtl = binary.BigEndian.Uint32(rawData[16:20])
	offset := 20

	bytesRead, err := r.NamePrimaryServer.Unmarshal(rawData[offset:])
	if err != nil {
		return offset, fmt.Errorf("unmarshalling NamePrimaryServer: %w", err)
	}
	offset += bytesRead

	bytesRead, err = r.ZoneAdminEmail.Unmarshal(rawData[offset:])
	if err != nil {
		return offset, fmt.Errorf("unmarshalling ZoneAdminEmail: %w", err)
	}
	offset += bytesRead

	return offset, nil
}
