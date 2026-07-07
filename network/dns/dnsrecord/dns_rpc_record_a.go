package dnsrecord

import (
	"fmt"
	"net"
)

// DNS_RPC_RECORD_A is the record-data payload for a DNS_TYPE_A record. It carries a single
// IPv4 address. In a dnsRecord attribute value the containing DNS_RECORD has a DataLength of 4
// and a Type of DNS_TYPE_A.
//
// Source: [MS-DNSP] DNS_RPC_RECORD_A (section 2.2.2.2.4.1), referenced from DNS_RPC_RECORD_DATA
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/38b87392-7f61-4e30-9263-9f0fb832d084
type DNS_RPC_RECORD_A struct {
	// IPAddress (4 bytes): An IPv4 address stored in network byte order.
	IPAddress [4]byte
}

// NewDNS_RPC_RECORD_A creates a new, empty DNS_RPC_RECORD_A.
//
// Returns:
// - A pointer to the new DNS_RPC_RECORD_A structure
func NewDNS_RPC_RECORD_A() *DNS_RPC_RECORD_A {
	return &DNS_RPC_RECORD_A{}
}

// SetIPv4 sets the address from a net.IP. The address MUST be an IPv4 address.
//
// Parameters:
// - ip: The IPv4 address to store
//
// Returns:
// - An error if ip is not an IPv4 address
func (r *DNS_RPC_RECORD_A) SetIPv4(ip net.IP) error {
	v4 := ip.To4()
	if v4 == nil {
		return fmt.Errorf("DNS_RPC_RECORD_A: %q is not an IPv4 address", ip.String())
	}
	copy(r.IPAddress[:], v4)
	return nil
}

// GetIPv4 returns the stored address as a net.IP.
//
// Returns:
// - The IPv4 address as a net.IP
func (r *DNS_RPC_RECORD_A) GetIPv4() net.IP {
	return net.IPv4(r.IPAddress[0], r.IPAddress[1], r.IPAddress[2], r.IPAddress[3])
}

// Marshal marshals the DNS_RPC_RECORD_A structure into a byte array.
//
// Returns:
// - A byte array representing the DNS_RPC_RECORD_A structure
// - An error if the marshaling fails
func (r *DNS_RPC_RECORD_A) Marshal() ([]byte, error) {
	marshalled := make([]byte, 4)
	copy(marshalled, r.IPAddress[:])
	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_RPC_RECORD_A structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read
// - An error if the unmarshaling fails
func (r *DNS_RPC_RECORD_A) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < 4 {
		return 0, fmt.Errorf("rawData too short for DNS_RPC_RECORD_A: need 4 bytes, have %d", len(rawData))
	}
	copy(r.IPAddress[:], rawData[:4])
	return 4, nil
}
