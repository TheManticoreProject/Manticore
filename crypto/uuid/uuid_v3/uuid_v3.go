package uuid_v3

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/uuid"
)

// UUIDv3 namespaces
// Source: https://www.rfc-editor.org/rfc/rfc4122#appendix-C
const (
	UUIDv3NamespaceDNS  = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	UUIDv3NamespaceURL  = "6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	UUIDv3NamespaceOID  = "6ba7b812-9dad-11d1-80b4-00c04fd430c8"
	UUIDv3NamespaceX500 = "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
)

// UUIDv3 represents a UUID v1 structure
//
// UUIDv3 is a structure that represents a UUID v1.
// It contains a UUID, a time, a clock sequence, and a node ID.
type UUIDv3 struct {
	uuid.UUID

	Namespace uuid.UUIDInterface

	Name string

	data [15]byte
}

// Marshal converts the UUIDv3 structure to a 16-byte array
//
// Returns:
//   - A byte slice containing the UUID's 16 bytes
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv3) Marshal() ([]byte, error) {
	// Contents of the UUIDv3 is the  MD5 hash of the namespace UUID + name
	if u.Namespace != nil {

		nsBytes, err := u.Namespace.Marshal()
		if err != nil {
			return nil, err
		}

		md5Hash := md5.New()
		md5Hash.Write(nsBytes)
		md5Hash.Write([]byte(u.Name))
		hashBytes := md5Hash.Sum(nil)

		hashBytesShifted := []byte{}

		hashBytesShifted = append(hashBytesShifted, hashBytes[0:6]...)

		a := hashBytes[6]<<4 | (hashBytes[7] >> 4)
		hashBytesShifted = append(hashBytesShifted, a)

		b := hashBytes[7]<<4 | (hashBytes[8] & 0xF)
		hashBytesShifted = append(hashBytesShifted, b)

		hashBytesShifted = append(hashBytesShifted, hashBytes[9:]...)

		// Copy the first 15 bytes of the hash to our data array
		copy(u.data[:], hashBytesShifted[0:15])

		fmt.Printf("hashBytesShifted: %x\n", hashBytesShifted)

	}

	// Set the UUID version and variant
	u.UUID.Version = 3
	u.UUID.Data = u.data

	// Use the UUID's Marshal method to get the final 16-byte array
	return u.UUID.Marshal()
}

// Unmarshal converts a 16-byte array into a UUIDv3 structure
//
// Returns:
//   - The number of bytes read
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv3) Unmarshal(marshalledData []byte) (int, error) {
	if len(marshalledData) < 16 {
		return 0, fmt.Errorf("invalid UUID length: got %d bytes, want 16 bytes", len(marshalledData))
	}

	copy(u.data[:], marshalledData[0:15])

	// First unmarshal into the generic UUID
	n, err := u.UUID.Unmarshal(marshalledData)
	if err != nil {
		return 0, err
	}

	// Check if this is a version 3 UUID
	if u.UUID.Version != 3 {
		return 0, fmt.Errorf("invalid UUID version: got %d, want 3", u.UUID.Version)
	}

	return n, nil
}

// FromString parses a UUID string in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
// and returns a UUIDv3 structure.
//
// Parameters:
//   - uuidStr: A string containing the UUID in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
//
// Returns:
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv3) FromString(uuidStr string) error {

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

// FromBytes creates a UUIDv3 from a 16-byte array
//
// Parameters:
//   - data: A byte slice containing the UUID's 16 bytes
//
// Returns:
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv3) FromBytes(data []byte) error {
	if len(data) != 16 {
		return fmt.Errorf("invalid UUID length: got %d bytes, want 16", len(data))
	}
	_, err := u.Unmarshal(data)
	return err
}

// String returns the string representation of the UUIDv3 structure
//
// Returns:
//   - The string representation of the UUIDv3 structure
func (u *UUIDv3) String() string {
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
