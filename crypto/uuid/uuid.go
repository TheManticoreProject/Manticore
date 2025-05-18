package uuid

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// UUID represents a UUID structure
type UUID struct {
	// Version is the version of the UUID
	Version uint8
	// Variant is the variant of the UUID
	Variant uint8

	// Data is the data of the UUID
	Data [15]byte
}

// Marshal converts a UUID structure into a 16-byte array
//
// Returns:
//   - A byte slice containing the UUID's 16 bytes
//   - An error if the UUID is invalid or the conversion fails
func (u *UUID) Marshal() ([]byte, error) {
	data := make([]byte, 16)

	// Copy the first 5 bytes of the UUID
	copy(data[0:6], u.Data[0:6])

	data6high := (u.Data[6] & 0xF0) >> 4
	data6low := u.Data[6] & 0x0F

	// Set the version in the 6th byte
	data[6] = data6high | (u.Version&0xF)<<4

	// Copy the 7th byte of the UUID (unchanged)
	data[7] = u.Data[7]

	// Set the variant in the 8th byte
	data[8] = data6low | (u.Variant&0xF)<<4

	// Copy the last 7 bytes of the UUID
	copy(data[9:16], u.Data[8:15])

	return data, nil
}

// Unmarshal converts a 16-byte array into a UUID structure
//
// Parameters:
//   - data: A byte slice containing the UUID's 16 bytes
//
// Returns:
//   - The number of bytes read
//   - An error if the UUID is invalid or the conversion fails
func (u *UUID) Unmarshal(data []byte) (int, error) {
	if len(data) < 16 {
		return 0, fmt.Errorf("invalid UUID length: got %d bytes, want 16 bytes", len(data))
	}
	// Copy the first 6 bytes of the UUID
	copy(u.Data[0:6], data[0:6])

	// Extract version from the 6th byte
	u.Version = (data[6] & 0xF0) >> 4
	data6high := data[6] & 0x0F

	// Copy the 7th byte
	u.Data[7] = data[7]

	// Extract variant from the 8th byte
	u.Variant = (data[8] & 0xF0) >> 4
	data6low := data[8] & 0x0F

	// Store the lower 4 bits of byte 8 in the upper 4 bits of byte 6
	u.Data[6] = (data6high << 4) | data6low

	// Copy the last 7 bytes (9-15)
	copy(u.Data[8:15], data[9:16])

	return 16, nil
}

// FromString parses a UUID string in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// and returns a UUID structure.
//
// Parameters:
//   - s: A string representation of the UUID
//
// Returns:
//   - An error if the string is not a valid UUID format
func (u *UUID) FromString(s string) error {
	// Remove hyphens and validate length
	s = strings.Replace(s, "-", "", -1)
	if len(s) != 32 {
		return fmt.Errorf("invalid UUID string length: got %d chars, want 32 chars", len(s))
	}

	data, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("invalid UUID string format: %v", err)
	}

	// Use Unmarshal to populate the UUID structure
	_, err = u.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("could not unmarshal UUID: %v", err)
	}

	return nil
}

// String returns a string representation of the UUID in the format
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx, where x is a hexadecimal digit.
//
// Returns:
//   - A string representation of the UUID
func (u *UUID) String() string {
	data, err := u.Marshal()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		data[0], data[1], data[2], data[3],
		data[4], data[5],
		data[6], data[7],
		data[8], data[9],
		data[10], data[11], data[12], data[13], data[14], data[15])
}
