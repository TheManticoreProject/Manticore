package key

import (
	"encoding/binary"
	"fmt"
)

// KeySource represents the source of the key.
// See: https://msdn.microsoft.com/en-us/library/mt220501.aspx
type KeySource int

const (
	// On Premises Key Trust
	KeySource_AD KeySource = 0x00

	// Hybrid Azure AD Key Trust
	KeySource_AzureAD KeySource = 0x01
)

// String returns the string representation of the KeySource.
//
// Returns:
// - A string representing the key source.
func (ks KeySource) String() string {
	switch ks {
	case KeySource_AD:
		return "Active Directory (AD)"
	case KeySource_AzureAD:
		return "Azure Active Directory (AAD)"
	default:
		return fmt.Sprintf("Unknown KeySource: %d", int(ks))
	}
}

// Parse parses the key source from a byte array.
//
// Parameters:
// - value: A byte array representing the key source.
//
// Returns:
// - A KeySource object.
func (ks KeySource) FromBytes(value []byte) KeySource {
	return KeySource(binary.LittleEndian.Uint16(value))
}
