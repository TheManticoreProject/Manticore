package msdnsp

import (
	"encoding/binary"
	"fmt"
)

// DNS_RPC_RECORD_NAME_PREFERENCE is the record-data payload for records that pair a 16-bit
// preference value with an FQDN. Per [MS-DNSP] it is used for the following record types:
//
//	DNS_TYPE_MX, DNS_TYPE_AFSDB, DNS_TYPE_RT
//
// WPreference is stored in network byte order (big-endian). NameExchange is stored in
// DNS_COUNT_NAME (section 2.2.2.2.2) form in a dnsRecord attribute value carried over LDAP.
//
// Source: [MS-DNSP] DNS_RPC_RECORD_NAME_PREFERENCE (section 2.2.2.2.4.8)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/f647d391-6614-4c3e-b38b-4df971590eb6
type DNS_RPC_RECORD_NAME_PREFERENCE struct {
	// WPreference (2 bytes): The preference value for the DNS server that holds the record.
	// Big-endian.
	WPreference uint16

	// NameExchange (variable): The FQDN of the server hosting the mail-exchange.
	NameExchange DNS_COUNT_NAME
}

// NewDNS_RPC_RECORD_NAME_PREFERENCE creates a new, empty DNS_RPC_RECORD_NAME_PREFERENCE.
//
// Returns:
// - A pointer to the new DNS_RPC_RECORD_NAME_PREFERENCE structure
func NewDNS_RPC_RECORD_NAME_PREFERENCE() *DNS_RPC_RECORD_NAME_PREFERENCE {
	return &DNS_RPC_RECORD_NAME_PREFERENCE{}
}

// Marshal marshals the DNS_RPC_RECORD_NAME_PREFERENCE structure into a byte array.
//
// Returns:
// - A byte array representing the DNS_RPC_RECORD_NAME_PREFERENCE structure
// - An error if the marshaling fails
func (r *DNS_RPC_RECORD_NAME_PREFERENCE) Marshal() ([]byte, error) {
	marshalled := make([]byte, 2)
	binary.BigEndian.PutUint16(marshalled[0:2], r.WPreference)

	exchange, err := r.NameExchange.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshalling NameExchange: %w", err)
	}
	marshalled = append(marshalled, exchange...)

	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_RPC_RECORD_NAME_PREFERENCE structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read
// - An error if the unmarshaling fails
func (r *DNS_RPC_RECORD_NAME_PREFERENCE) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < 2 {
		return 0, fmt.Errorf("rawData too short for DNS_RPC_RECORD_NAME_PREFERENCE fixed field: need 2 bytes, have %d", len(rawData))
	}

	r.WPreference = binary.BigEndian.Uint16(rawData[0:2])
	offset := 2

	bytesRead, err := r.NameExchange.Unmarshal(rawData[offset:])
	if err != nil {
		return offset, fmt.Errorf("unmarshalling NameExchange: %w", err)
	}
	offset += bytesRead

	return offset, nil
}
