package version

import (
	"encoding/binary"
	"fmt"
)

type KeyCredentialLinkVersion struct {
	Value uint32

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

const (
	KeyCredentialLinkVersion_0 uint32 = 0x0
	KeyCredentialLinkVersion_1 uint32 = 0x00000100
	KeyCredentialLinkVersion_2 uint32 = 0x00000200
)

// Unmarshal parses the KeyCredentialLinkVersion from a byte array.
//
// Parameters:
// - value: A byte array representing the KeyCredentialLinkVersion.
//
// Returns:
// - The number of bytes read (4 on success).
// - An error if value is shorter than 4 bytes.
func (kcv *KeyCredentialLinkVersion) Unmarshal(value []byte) (int, error) {
	if len(value) < 4 {
		return 0, fmt.Errorf("data too short to unmarshal KeyCredentialLinkVersion: have %d bytes, need 4", len(value))
	}

	kcv.RawBytes = value[:4]
	kcv.RawBytesSize = 4

	kcv.Value = binary.LittleEndian.Uint32(value[:4])

	return 4, nil
}

// Marshal returns the raw bytes of the KeyCredentialLinkVersion structure.
//
// Returns:
// - A byte slice representing the raw bytes of the KeyCredentialLinkVersion structure.
func (kcv *KeyCredentialLinkVersion) Marshal() ([]byte, error) {
	buffer := make([]byte, 4)
	binary.LittleEndian.PutUint32(buffer, kcv.Value)
	return buffer, nil
}

// String returns a string representation of the KeyCredentialLinkVersion.
//
// Returns:
// - A string representing the KeyCredentialLinkVersion.
func (kcv *KeyCredentialLinkVersion) String() string {
	switch kcv.Value {
	case KeyCredentialLinkVersion_0:
		return "KeyCredentialLink_v0"
	case KeyCredentialLinkVersion_1:
		return "KeyCredentialLink_v1"
	case KeyCredentialLinkVersion_2:
		return "KeyCredentialLink_v2"
	}

	return fmt.Sprintf("Unknown version: 0x%08x", kcv.Value)
}
