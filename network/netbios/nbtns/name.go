package nbtns

import (
	"bytes"
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
