package uuid_v8

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/uuid"
)

// UUIDv8 namespaces
// Source: https://www.rfc-editor.org/rfc/rfc4122#appendix-C
const (
	UUIDv8NamespaceDNS  = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	UUIDv8NamespaceURL  = "6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	UUIDv8NamespaceOID  = "6ba7b812-9dad-11d1-80b4-00c04fd430c8"
	UUIDv8NamespaceX500 = "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
)

// UUIDv8 represents a UUID v1 structure
//
// UUIDv8 is a structure that represents a UUID v1.
// It contains a UUID, a time, a clock sequence, and a node ID.
type UUIDv8 struct {
	uuid.UUID

	Data [15]byte
}

// Marshal converts the UUIDv8 structure to a 16-byte array
//
// Returns:
//   - A byte slice containing the UUID's 16 bytes
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv8) Marshal() ([]byte, error) {
	// Set the UUID version and variant
	u.UUID.Version = 8
	u.UUID.Data = u.Data

	// Use the UUID's Marshal method to get the final 16-byte array
	return u.UUID.Marshal()
}

// Unmarshal converts a 16-byte array into a UUIDv8 structure
//
// Returns:
//   - The number of bytes read
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv8) Unmarshal(marshalledData []byte) (int, error) {
	if len(marshalledData) < 16 {
		return 0, fmt.Errorf("invalid UUID length: got %d bytes, want 16 bytes", len(marshalledData))
	}

	// First unmarshal into the generic UUID
	n, err := u.UUID.Unmarshal(marshalledData)
	if err != nil {
		return 0, err
	}

	copy(u.Data[:], u.UUID.Data[:])

	// Check if this is a version 8 UUID
	if u.UUID.Version != 8 {
		return 0, fmt.Errorf("invalid UUID version: got %d, want 8", u.UUID.Version)
	}

	return n, nil
}

// FromString parses a UUID string in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
// and returns a UUIDv8 structure.
//
// Parameters:
//   - uuidStr: A string containing the UUID in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
//
// Returns:
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv8) FromString(uuidStr string) error {

	uuidStr = strings.ReplaceAll(uuidStr, "-", "")

	if len(uuidStr) != 32 {
		return fmt.Errorf("invalid UUID length: got %d characters, want 32", len(uuidStr))
	}

	uuidStr = strings.ToLower(uuidStr)

	uuidBytes, err := hex.DecodeString(uuidStr)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	return u.FromBytes(uuidBytes)
}

// FromBytes creates a UUIDv8 from a 16-byte array
//
// Parameters:
//   - data: A byte slice containing the UUID's 16 bytes
//
// Returns:
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv8) FromBytes(data []byte) error {
	if len(data) != 16 {
		return fmt.Errorf("invalid UUID length: got %d bytes, want 16", len(data))
	}
	_, err := u.Unmarshal(data)
	return err
}

// SetData sets the data of the UUIDv8 structure
//
// Parameters:
//   - data: A byte slice containing the UUID's 15 bytes
func (u *UUIDv8) SetData(data []byte) {
	copy(u.Data[:], data[:])
}

// GetData returns the data of the UUIDv8 structure
//
// Returns:
//   - A byte slice containing the UUID's 15 bytes
func (u *UUIDv8) GetData() []byte {
	return u.Data[:]
}

// String returns the string representation of the UUIDv8 structure
//
// Returns:
//   - The string representation of the UUIDv8 structure
func (u *UUIDv8) String() string {
	// Use the UUID's Marshal method to get the 16-byte array
	marshalledData, err := u.Marshal()
	if err != nil {
		return fmt.Sprintf("invalid UUID: %v", err)
	}

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(marshalledData[0:4]),
		binary.BigEndian.Uint16(marshalledData[4:6]),
		binary.BigEndian.Uint16(marshalledData[6:8]),
		binary.BigEndian.Uint16(marshalledData[8:10]),
		marshalledData[10:16],
	)
}
