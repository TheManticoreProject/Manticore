package key

import (
	"encoding/binary"
	"fmt"
)

type KeyCredentialVersion struct {
	Value uint32

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

const (
	KeyCredentialVersion_0 uint32 = 0x0
	KeyCredentialVersion_1 uint32 = 0x00000100
	KeyCredentialVersion_2 uint32 = 0x00000200
)

// Parse parses the KeyCredentialVersion from a byte array.
//
// Parameters:
// - value: A byte array representing the KeyCredentialVersion.
func (kcv *KeyCredentialVersion) FromBytes(value []byte) {
	kcv.RawBytes = value[:4]
	kcv.RawBytesSize = 4

	kcv.Value = binary.LittleEndian.Uint32(value[:4])
}

// String returns a string representation of the KeyCredentialVersion.
//
// Returns:
// - A string representing the KeyCredentialVersion.
func (kcv *KeyCredentialVersion) String() string {
	switch kcv.Value {
	case KeyCredentialVersion_0:
		return "KeyCredential_v0"
	case KeyCredentialVersion_1:
		return "KeyCredential_v1"
	case KeyCredentialVersion_2:
		return "KeyCredential_v2"
	}

	return fmt.Sprintf("Unknown version: %d", int(kcv.Value))
}

// ToBytes returns the raw bytes of the KeyCredentialVersion structure.
//
// Returns:
// - A byte slice representing the raw bytes of the KeyCredentialVersion structure.
func (kcv *KeyCredentialVersion) ToBytes() []byte {
	buffer := make([]byte, 4)
	binary.LittleEndian.PutUint32(buffer, kcv.Value)
	return buffer
}
