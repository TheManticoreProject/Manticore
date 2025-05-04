package key

import (
	"fmt"
)

/*
Key Credential Link Entry Identifier

Describes the data stored in the Value field.
https://msdn.microsoft.com/en-us/library/mt220499.aspx
*/
type KeyCredentialEntryType struct {
	Value uint8

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

const (
	// A SHA256 hash of the Value field of the KeyMaterial entry.
	KeyCredentialEntryType_KeyID uint8 = 0x01

	// A SHA256 hash of all entries following this entry.
	KeyCredentialEntryType_KeyHash uint8 = 0x02

	// Key material of the credential.
	KeyCredentialEntryType_KeyMaterial uint8 = 0x03

	// Key Usage
	KeyCredentialEntryType_KeyUsage uint8 = 0x04

	// Key Source
	KeyCredentialEntryType_KeySource uint8 = 0x05

	// Device Identifier
	KeyCredentialEntryType_DeviceId uint8 = 0x06

	// Custom key information.
	KeyCredentialEntryType_CustomKeyInformation uint8 = 0x07

	// The approximate time this key was last used, in FILETIME format.
	KeyCredentialEntryType_KeyApproximateLastLogonTimeStamp uint8 = 0x08

	// The approximate time this key was created, in FILETIME format.
	KeyCredentialEntryType_KeyCreationTime uint8 = 0x09
)

func (k *KeyCredentialEntryType) FromBytes(value byte) {
	k.RawBytes = []byte{value}
	k.RawBytesSize = 1

	k.Value = value
}

// String returns a string representation of the KeyCredentialEntryType.
//
// Returns:
// - A string representing the KeyCredentialEntryType.
func (k *KeyCredentialEntryType) String() string {
	switch k.Value {
	case KeyCredentialEntryType_KeyID:
		return "KeyID"
	case KeyCredentialEntryType_KeyHash:
		return "KeyHash"
	case KeyCredentialEntryType_KeyMaterial:
		return "KeyMaterial"
	case KeyCredentialEntryType_KeyUsage:
		return "KeyUsage"
	case KeyCredentialEntryType_KeySource:
		return "KeySource"
	case KeyCredentialEntryType_DeviceId:
		return "DeviceId"
	case KeyCredentialEntryType_CustomKeyInformation:
		return "CustomKeyInformation"
	case KeyCredentialEntryType_KeyApproximateLastLogonTimeStamp:
		return "KeyApproximateLastLogonTimeStamp"
	case KeyCredentialEntryType_KeyCreationTime:
		return "KeyCreationTime"
	default:
		return fmt.Sprintf("Unknown KeyCredentialEntryType: %d", k.Value)
	}
}

// ToBytes returns the raw bytes of the KeyCredentialEntryType structure.
//
// Returns:
// - A byte slice representing the raw bytes of the KeyCredentialEntryType structure.
func (k *KeyCredentialEntryType) ToBytes() []byte {
	return []byte{k.Value}
}
