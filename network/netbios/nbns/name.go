package nbns

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

// NameType indicates whether a name is unique or group
type NameType uint8

const (
	Unique NameType = iota // Only one owner allowed
	Group                  // Multiple owners allowed
)

// NameStatus represents the current state of a name
type NameStatus uint8

const (
	Active NameStatus = iota
	Conflict
	Releasing
)

// NameRecord represents a registered NetBIOS name and its attributes
type NameRecord struct {
	Name            string
	Type            NameType
	Status          NameStatus
	Owners          []net.IP  // IP addresses of nodes that own this name
	TTL             time.Time // Time-to-live for name registration
	RefreshInterval time.Duration
	ScopeID         string // NetBIOS scope identifier
}

// NetBIOSName represents a NetBIOS name with its scope
type NetBIOSName struct {
	Name    string
	ScopeID string
}

// POSITIVE NAME QUERY RESPONSE
type ADDR_ENTRY struct {
	Flags   uint16
	Address uint32
}

// NB_FLAGS bits carried in the ADDR_ENTRY.Flags field of an NB resource record
// (RFC 1002 4.2.1.3: G | ONT | RESERVED). The Group (G) bit is the most
// significant bit of the 16-bit NB_FLAGS field, distinct from the header flags.
const (
	NBFlagGroup uint16 = 0x8000 // G: set for a group name, clear for a unique name
)

// Constants for name encoding
const (
	NetBIOSNameLength = 16 // NetBIOS names are exactly 16 bytes
	EncodedNameLength = 32 // Each half-byte becomes a byte in encoded form
	ASCII_A           = 0x41
)

// Validate checks if a NetBIOS name is valid
func (n *NetBIOSName) Validate() error {
	if len(n.Name) > NetBIOSNameLength {
		return fmt.Errorf("name too long: max %d bytes", NetBIOSNameLength)
	}

	// NetBIOS names cannot start with *
	if strings.HasPrefix(n.Name, "*") {
		return fmt.Errorf("name cannot start with *")
	}

	// Validate scope ID format if present
	if n.ScopeID != "" {
		if !isValidDomainName(n.ScopeID) {
			return fmt.Errorf("invalid scope ID format")
		}
	}

	return nil
}

// FirstLevelEncode performs the first level encoding of a NetBIOS name
// as specified in RFC 1001 section 14.1
func (n *NetBIOSName) FirstLevelEncode() (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}

	// Pad name to exactly 16 bytes with spaces
	name := make([]byte, NetBIOSNameLength)
	copy(name, n.Name)
	for i := len(n.Name); i < NetBIOSNameLength; i++ {
		name[i] = ' '
	}

	// Encode each half-byte into a byte
	encoded := make([]byte, EncodedNameLength)
	for i := 0; i < NetBIOSNameLength; i++ {
		encoded[i*2] = ((name[i] >> 4) & 0x0F) + ASCII_A
		encoded[i*2+1] = (name[i] & 0x0F) + ASCII_A
	}

	// Add scope ID if present
	result := string(encoded)
	if n.ScopeID != "" {
		result = result + "." + n.ScopeID
	}

	return result, nil
}

// EncodeSessionServiceName encodes name (with the given one-byte service suffix)
// into the 34-byte "second-level" form the NetBIOS session service carries in a
// SESSION REQUEST (RFC 1002 4.3.2 / RFC 1001 14.1): a 0x20 length byte, the
// 32-byte first-level encoding of the 16-byte name (up to 15 characters padded
// with spaces plus the suffix byte in the final position), and a single 0x00
// label terminator (no scope for the default). Unlike FirstLevelEncode it
// permits a leading '*' wildcard (e.g. the "*SMBSERVER" convention), which the
// session service requires and which Validate would otherwise reject.
func EncodeSessionServiceName(name string, suffix byte) ([]byte, error) {
	if len(name) > NetBIOSNameLength-1 {
		return nil, fmt.Errorf("name too long: max %d bytes", NetBIOSNameLength-1)
	}

	// Build the 16-byte name: up to 15 characters padded with spaces, with the
	// service suffix occupying the final byte.
	name16 := make([]byte, NetBIOSNameLength)
	copy(name16, name)
	for i := len(name); i < NetBIOSNameLength-1; i++ {
		name16[i] = ' '
	}
	name16[NetBIOSNameLength-1] = suffix

	// Second-level encoding: the 0x20 length byte, the first-level encoding
	// (each half-byte mapped to a printable character by adding 'A'), then the
	// 0x00 label terminator.
	encoded := make([]byte, 0, EncodedNameLength+2)
	encoded = append(encoded, EncodedNameLength)
	for i := 0; i < NetBIOSNameLength; i++ {
		encoded = append(encoded, ((name16[i]>>4)&0x0F)+ASCII_A)
		encoded = append(encoded, (name16[i]&0x0F)+ASCII_A)
	}
	encoded = append(encoded, 0x00)

	return encoded, nil
}

// DecodeSessionServiceName decodes one second-level-encoded NetBIOS name as it
// appears in a SESSION REQUEST (RFC 1002 4.3.2), the inverse of
// EncodeSessionServiceName: a 0x20 length byte, EncodedNameLength encoded bytes,
// then the label sequence terminating the name — a bare 0x00 for the default
// scope, or one or more length-prefixed scope labels followed by 0x00.
//
// It returns the name with its trailing space padding removed, the one-byte
// service suffix that occupied the 16th position, and the total number of bytes
// consumed, so a caller can decode the CALLED name and then the CALLING name
// that follows it.
//
// Parameters:
//   - b: the buffer positioned at the start of the encoded name
//
// Returns:
//   - name: the decoded name, trailing padding removed
//   - suffix: the one-byte service suffix
//   - n: the number of bytes consumed from b
//   - err: non-nil if the encoding is malformed
func DecodeSessionServiceName(b []byte) (name string, suffix byte, n int, err error) {
	// The shortest well-formed name is the length byte, the encoded name and a
	// single 0x00 label terminator.
	if len(b) < 1+EncodedNameLength+1 {
		return "", 0, 0, fmt.Errorf("truncated session-service name: need at least %d bytes, got %d", 1+EncodedNameLength+1, len(b))
	}
	if b[0] != EncodedNameLength {
		return "", 0, 0, fmt.Errorf("invalid session-service name length byte: expected 0x%02X, got 0x%02X", EncodedNameLength, b[0])
	}

	// First-level decoding: each pair of characters in the range 'A'..'P' encodes
	// the high and low half-byte of one name byte.
	decoded := make([]byte, NetBIOSNameLength)
	for i := 0; i < NetBIOSNameLength; i++ {
		high := b[1+i*2] - ASCII_A
		low := b[1+i*2+1] - ASCII_A
		if high > 0x0F || low > 0x0F {
			return "", 0, 0, fmt.Errorf("invalid encoding character at offset %d", 1+i*2)
		}
		decoded[i] = (high << 4) | low
	}

	n = 1 + EncodedNameLength

	// Walk the remaining labels. The default scope is a single 0x00 terminator;
	// a scoped name carries length-prefixed labels before it. Bounding each
	// label by the DNS 63-byte limit keeps a malformed buffer from being read as
	// one enormous label.
	for {
		if n >= len(b) {
			return "", 0, 0, fmt.Errorf("session-service name is not terminated")
		}
		labelLength := int(b[n])
		n++
		if labelLength == 0 {
			break
		}
		if labelLength > 63 {
			return "", 0, 0, fmt.Errorf("invalid scope label length %d at offset %d", labelLength, n-1)
		}
		if n+labelLength > len(b) {
			return "", 0, 0, fmt.Errorf("truncated scope label at offset %d", n-1)
		}
		n += labelLength
	}

	// The 16th byte is the service suffix; the preceding 15 are the name, padded
	// with spaces on the wire.
	suffix = decoded[NetBIOSNameLength-1]
	name = string(bytes.TrimRight(decoded[:NetBIOSNameLength-1], " "))

	return name, suffix, n, nil
}

// FirstLevelDecode decodes a first level encoded NetBIOS name
func FirstLevelDecode(encoded string) (*NetBIOSName, error) {
	parts := strings.SplitN(encoded, ".", 2)
	encodedName := parts[0]

	if len(encodedName) != EncodedNameLength {
		return nil, fmt.Errorf("invalid encoded name length")
	}

	// Decode each pair of bytes back into a single byte
	decoded := make([]byte, NetBIOSNameLength)
	for i := 0; i < NetBIOSNameLength; i++ {
		high := encodedName[i*2] - ASCII_A
		low := encodedName[i*2+1] - ASCII_A

		if high > 0x0F || low > 0x0F {
			return nil, fmt.Errorf("invalid encoding character")
		}

		decoded[i] = (high << 4) | low
	}

	// Trim trailing spaces
	name := string(bytes.TrimRight(decoded, " "))

	nbName := &NetBIOSName{
		Name: name,
	}

	// Add scope ID if present
	if len(parts) > 1 {
		nbName.ScopeID = parts[1]
	}

	return nbName, nil
}

// isValidDomainName checks if a string is a valid domain name
func isValidDomainName(name string) bool {
	// Simple validation - could be made more thorough
	if name == "" {
		return false
	}

	parts := strings.Split(name, ".")
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}

		for _, c := range part {
			if !(c >= 'a' && c <= 'z' ||
				c >= 'A' && c <= 'Z' ||
				c >= '0' && c <= '9' ||
				c == '-') {
				return false
			}
		}

		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}
	}

	return true
}

// Marshals ADDR_ENTRY to buffer
func (n *ADDR_ENTRY) Marshal() []byte {
	buf := make([]byte, 6)
	binary.BigEndian.PutUint16(buf[0:2], n.Flags)
	binary.BigEndian.PutUint32(buf[2:6], n.Address)
	return buf
}

// Unmarshal decodes an ADDR_ENTRY from a byte slice
func (n *ADDR_ENTRY) Unmarshal(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("ADDR_ENTRY too short: need 6 bytes, got %d", len(data))
	}
	n.Flags = binary.BigEndian.Uint16(data[0:2])
	n.Address = binary.BigEndian.Uint32(data[2:6])
	return nil
}

// IP returns the address as a net.IP
func (n *ADDR_ENTRY) IP() net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n.Address)
	return ip
}

// ParseIPFromRData extracts an IP address from NB resource record RData.
// RData is expected to be in ADDR_ENTRY format (2 bytes flags + 4 bytes address).
func ParseIPFromRData(rdata []byte) (net.IP, error) {
	var entry ADDR_ENTRY
	if err := entry.Unmarshal(rdata); err != nil {
		return nil, err
	}
	return entry.IP(), nil
}

// returns static 6 bytes number
func (n *ADDR_ENTRY) Length() uint16 {
	return 6
}
