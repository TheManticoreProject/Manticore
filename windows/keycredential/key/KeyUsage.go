package key

import "fmt"

type KeyUsage struct {
	Value uint8

	// Internal
	RawBytes     []byte
	RawBytesSize uint8
}

const (
	// Key Usage
	// See: https://msdn.microsoft.com/en-us/library/mt220501.aspx

	// Admin key (pin-reset key)
	KeyUsage_AdminKey uint8 = 0

	// NGC key attached to a user object (KEY_USAGE_NGC)
	KeyUsage_NGC uint8 = 0x01

	// Transport key attached to a device object
	KeyUsage_STK uint8 = 0x02

	// BitLocker recovery key
	KeyUsage_BitlockerRecovery uint8 = 0x03

	// Unrecognized key usage
	KeyUsage_Other uint8 = 0x04

	// Fast IDentity Online Key (KEY_USAGE_FIDO)
	KeyUsage_FIDO uint8 = 0x07

	// File Encryption Key (KEY_USAGE_FEK)
	KeyUsage_FEK uint8 = 0x08

	// DPAPI Key
	// TODO: The DPAPI enum needs to be mapped to a proper integer value.
	KeyUsage_DPAPI uint8 = 0x09
)

// Parse parses the key usage from a byte array.
//
// Parameters:
// - value: A byte array representing the key usage.
func (ku *KeyUsage) FromBytes(value byte) {
	ku.RawBytes = []byte{value}
	ku.RawBytesSize = 1

	ku.Value = value
}

// String returns a string representation of the key usage.
//
// Returns:
// - A string representing the key usage.
func (ku *KeyUsage) String() string {
	switch ku.Value {
	case KeyUsage_AdminKey:
		return "AdminKey"
	case KeyUsage_NGC:
		return "New Generation Credential (NGC)"
	case KeyUsage_STK:
		return "Smart Token Key (STK)"
	case KeyUsage_BitlockerRecovery:
		return "Bitlocker Recovery"
	case KeyUsage_Other:
		return "Other"
	case KeyUsage_FIDO:
		return "Fast IDentity Online (FIDO)"
	case KeyUsage_FEK:
		return "File Encryption Key (FEK)"
	case KeyUsage_DPAPI:
		return "Data Protection API (DPAPI)"
	}

	return fmt.Sprintf("Unknown KeyUsage: %d", ku.Value)
}
