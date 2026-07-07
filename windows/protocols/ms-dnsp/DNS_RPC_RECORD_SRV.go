package msdnsp

import (
	"encoding/binary"
	"fmt"
)

// DNS_RPC_RECORD_SRV is the record-data payload for a DNS_TYPE_SRV record, as specified in
// [RFC2782].
//
// The three numeric fields are stored in network byte order (big-endian). NameTarget is
// stored in DNS_COUNT_NAME (section 2.2.2.2.2) form in a dnsRecord attribute value carried
// over LDAP.
//
// Source: [MS-DNSP] DNS_RPC_RECORD_SRV (section 2.2.2.2.4.7)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/db37cab7-f121-43ba-81c5-ca0e198d4b9a
type DNS_RPC_RECORD_SRV struct {
	// WPriority (2 bytes): The priority of the target host, per [RFC2782]. Big-endian.
	WPriority uint16

	// WWeight (2 bytes): The relative weight for the target host, per [RFC2782]. Big-endian.
	WWeight uint16

	// WPort (2 bytes): The port number for the service on the target host, per [RFC2782].
	// Big-endian.
	WPort uint16

	// NameTarget (variable): The FQDN of the server that hosts this service.
	NameTarget DNS_COUNT_NAME
}

// NewDNS_RPC_RECORD_SRV creates a new, empty DNS_RPC_RECORD_SRV.
//
// Returns:
// - A pointer to the new DNS_RPC_RECORD_SRV structure
func NewDNS_RPC_RECORD_SRV() *DNS_RPC_RECORD_SRV {
	return &DNS_RPC_RECORD_SRV{}
}

// Marshal marshals the DNS_RPC_RECORD_SRV structure into a byte array.
//
// Returns:
// - A byte array representing the DNS_RPC_RECORD_SRV structure
// - An error if the marshaling fails
func (r *DNS_RPC_RECORD_SRV) Marshal() ([]byte, error) {
	marshalled := make([]byte, 6)
	binary.BigEndian.PutUint16(marshalled[0:2], r.WPriority)
	binary.BigEndian.PutUint16(marshalled[2:4], r.WWeight)
	binary.BigEndian.PutUint16(marshalled[4:6], r.WPort)

	target, err := r.NameTarget.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshalling NameTarget: %w", err)
	}
	marshalled = append(marshalled, target...)

	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_RPC_RECORD_SRV structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read
// - An error if the unmarshaling fails
func (r *DNS_RPC_RECORD_SRV) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < 6 {
		return 0, fmt.Errorf("rawData too short for DNS_RPC_RECORD_SRV fixed fields: need 6 bytes, have %d", len(rawData))
	}

	r.WPriority = binary.BigEndian.Uint16(rawData[0:2])
	r.WWeight = binary.BigEndian.Uint16(rawData[2:4])
	r.WPort = binary.BigEndian.Uint16(rawData[4:6])
	offset := 6

	bytesRead, err := r.NameTarget.Unmarshal(rawData[offset:])
	if err != nil {
		return offset, fmt.Errorf("unmarshalling NameTarget: %w", err)
	}
	offset += bytesRead

	return offset, nil
}
