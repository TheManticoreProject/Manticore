package msdnsp

import (
	"fmt"
	"net"
)

// DNS_RPC_RECORD_AAAA is the record-data payload for a DNS_TYPE_AAAA record. It carries a
// single IPv6 address. In a dnsRecord attribute value the containing DNS_RECORD has a
// DataLength of 16 and a Type of DNS_TYPE_AAAA.
//
// Source: [MS-DNSP] DNS_RPC_RECORD_AAAA (section 2.2.2.2.4.17)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/ee33fef1-6e82-42d0-8107-0f6d21be072a
type DNS_RPC_RECORD_AAAA struct {
	// IPv6Address (16 bytes): An IPv6 address stored in network byte order.
	IPv6Address [16]byte
}

// NewDNS_RPC_RECORD_AAAA creates a new, empty DNS_RPC_RECORD_AAAA.
//
// Returns:
// - A pointer to the new DNS_RPC_RECORD_AAAA structure
func NewDNS_RPC_RECORD_AAAA() *DNS_RPC_RECORD_AAAA {
	return &DNS_RPC_RECORD_AAAA{}
}

// SetIPv6 sets the address from a net.IP. The address MUST be representable as a 16-byte IPv6
// address.
//
// Parameters:
// - ip: The IPv6 address to store
//
// Returns:
// - An error if ip cannot be represented as a 16-byte address
func (r *DNS_RPC_RECORD_AAAA) SetIPv6(ip net.IP) error {
	v6 := ip.To16()
	if v6 == nil {
		return fmt.Errorf("DNS_RPC_RECORD_AAAA: %q is not a valid IP address", ip.String())
	}
	copy(r.IPv6Address[:], v6)
	return nil
}

// GetIPv6 returns the stored address as a net.IP.
//
// Returns:
// - The IPv6 address as a net.IP
func (r *DNS_RPC_RECORD_AAAA) GetIPv6() net.IP {
	ip := make(net.IP, 16)
	copy(ip, r.IPv6Address[:])
	return ip
}

// Marshal marshals the DNS_RPC_RECORD_AAAA structure into a byte array.
//
// Returns:
// - A byte array representing the DNS_RPC_RECORD_AAAA structure
// - An error if the marshaling fails
func (r *DNS_RPC_RECORD_AAAA) Marshal() ([]byte, error) {
	marshalled := make([]byte, 16)
	copy(marshalled, r.IPv6Address[:])
	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_RPC_RECORD_AAAA structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read
// - An error if the unmarshaling fails
func (r *DNS_RPC_RECORD_AAAA) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < 16 {
		return 0, fmt.Errorf("rawData too short for DNS_RPC_RECORD_AAAA: need 16 bytes, have %d", len(rawData))
	}
	copy(r.IPv6Address[:], rawData[:16])
	return 16, nil
}
