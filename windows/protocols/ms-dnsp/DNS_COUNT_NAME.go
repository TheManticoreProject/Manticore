package msdnsp

import (
	"fmt"
	"strings"
)

// DNS_COUNT_NAME is the on-directory form of an FQDN used inside dnsRecord attribute values
// carried over LDAP. When a dnsRecord value is written to the directory each name is
// converted from DNS_RPC_NAME (section 2.2.2.2.1) to DNS_COUNT_NAME, and converted back when
// read.
//
// The name is encoded as a sequence of length-prefixed labels: a 1-byte label-length count is
// inserted before the first label and in place of each "." delimiter, and the sequence is
// null-terminated. For example "example.com" is encoded as 07 'example' 03 'com' 00, with a
// Length of 13 and a LabelCount of 2.
//
// Source: [MS-DNSP] DNS_COUNT_NAME (section 2.2.2.2.2)
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dnsp/2af86306-3bd9-4c86-b763-fd33e16e3e5b
type DNS_COUNT_NAME struct {
	// Length (1 byte): The length, in bytes, of the string stored in the RawName member,
	// including null termination. To represent an empty string, Length MUST be zero,
	// LabelCount MUST be zero, and RawName MUST be empty.
	Length uint8

	// LabelCount (1 byte): The count of DNS labels in the RawName member.
	LabelCount uint8

	// RawName (variable): A string containing an FQDN in which a 1-byte label length count
	// for the subsequent label has been inserted before the first label and in place of each
	// "." delimiter. The string MUST be null-terminated. The maximum length of the string,
	// including the null terminator, is 256 bytes.
	RawName []byte
}

// NewDNS_COUNT_NAME creates a new, empty DNS_COUNT_NAME.
//
// Returns:
// - A pointer to the new DNS_COUNT_NAME structure
func NewDNS_COUNT_NAME() *DNS_COUNT_NAME {
	return &DNS_COUNT_NAME{}
}

// NewDNS_COUNT_NAMEFromFQDN creates a DNS_COUNT_NAME from a dotted FQDN string.
//
// Parameters:
// - fqdn: The dotted fully-qualified domain name (a trailing "." is tolerated and ignored)
//
// Returns:
//   - A pointer to the populated DNS_COUNT_NAME structure
//   - An error if any label is empty or exceeds 63 bytes, or if the encoded name exceeds
//     the 256-byte maximum
func NewDNS_COUNT_NAMEFromFQDN(fqdn string) (*DNS_COUNT_NAME, error) {
	n := &DNS_COUNT_NAME{}
	if err := n.SetFQDN(fqdn); err != nil {
		return nil, err
	}
	return n, nil
}

// SetFQDN encodes the provided dotted FQDN into the RawName, Length, and LabelCount fields.
//
// Parameters:
// - fqdn: The dotted fully-qualified domain name (a single trailing "." is tolerated)
//
// Returns:
//   - An error if any label is empty or longer than 63 bytes, or if the encoded name
//     (including the null terminator) exceeds 255 bytes.
func (n *DNS_COUNT_NAME) SetFQDN(fqdn string) error {
	// A single trailing dot denotes the root and is not a label separator.
	fqdn = strings.TrimSuffix(fqdn, ".")

	if fqdn == "" {
		n.Length = 0
		n.LabelCount = 0
		n.RawName = nil
		return nil
	}

	labels := strings.Split(fqdn, ".")
	raw := make([]byte, 0, len(fqdn)+2)
	for _, label := range labels {
		if len(label) == 0 {
			return fmt.Errorf("invalid FQDN %q: empty label", fqdn)
		}
		if len(label) > 63 {
			return fmt.Errorf("invalid FQDN %q: label %q exceeds 63 bytes", fqdn, label)
		}
		raw = append(raw, byte(len(label)))
		raw = append(raw, []byte(label)...)
	}
	// Null terminator.
	raw = append(raw, 0x00)

	// The spec states a 256-byte maximum including the terminator, but the Length field is a
	// single byte, so the representable maximum is 255. Cap here to avoid overflowing Length.
	if len(raw) > 255 {
		return fmt.Errorf("invalid FQDN %q: encoded name is %d bytes, exceeds 255", fqdn, len(raw))
	}

	n.RawName = raw
	n.Length = uint8(len(raw))
	n.LabelCount = uint8(len(labels))
	return nil
}

// GetFQDN decodes the RawName into a dotted FQDN string. An empty name decodes to "".
//
// Returns:
// - The dotted FQDN
// - An error if the RawName is malformed (a label length runs past the end of the buffer)
func (n *DNS_COUNT_NAME) GetFQDN() (string, error) {
	if len(n.RawName) == 0 {
		return "", nil
	}

	labels := make([]string, 0, n.LabelCount)
	offset := 0
	for i := 0; i < int(n.LabelCount); i++ {
		if offset >= len(n.RawName) {
			return "", fmt.Errorf("malformed DNS_COUNT_NAME: label %d begins past end of RawName", i)
		}
		labelLen := int(n.RawName[offset])
		offset++
		if offset+labelLen > len(n.RawName) {
			return "", fmt.Errorf("malformed DNS_COUNT_NAME: label %d length %d runs past end of RawName", i, labelLen)
		}
		labels = append(labels, string(n.RawName[offset:offset+labelLen]))
		offset += labelLen
	}

	return strings.Join(labels, "."), nil
}

// Marshal marshals the DNS_COUNT_NAME structure into a byte array.
//
// Returns:
// - A byte array representing the DNS_COUNT_NAME structure
// - An error if the marshaling fails
func (n *DNS_COUNT_NAME) Marshal() ([]byte, error) {
	if len(n.RawName) > 255 {
		return nil, fmt.Errorf("DNS_COUNT_NAME RawName is %d bytes, cannot exceed 255", len(n.RawName))
	}

	marshalled := make([]byte, 0, 2+len(n.RawName))
	marshalled = append(marshalled, n.Length)
	marshalled = append(marshalled, n.LabelCount)
	marshalled = append(marshalled, n.RawName...)
	return marshalled, nil
}

// Unmarshal unmarshals a byte array into the DNS_COUNT_NAME structure.
//
// Parameters:
// - rawData: The byte array to unmarshal
//
// Returns:
// - The number of bytes read
// - An error if the unmarshaling fails
func (n *DNS_COUNT_NAME) Unmarshal(rawData []byte) (int, error) {
	if len(rawData) < 2 {
		return 0, fmt.Errorf("rawData too short for DNS_COUNT_NAME header: %d bytes", len(rawData))
	}

	n.Length = rawData[0]
	n.LabelCount = rawData[1]

	offset := 2
	if len(rawData) < offset+int(n.Length) {
		return 0, fmt.Errorf("rawData too short for DNS_COUNT_NAME RawName: need %d bytes, have %d", offset+int(n.Length), len(rawData))
	}

	if n.Length == 0 {
		n.RawName = nil
	} else {
		n.RawName = make([]byte, n.Length)
		copy(n.RawName, rawData[offset:offset+int(n.Length)])
	}
	offset += int(n.Length)

	return offset, nil
}
