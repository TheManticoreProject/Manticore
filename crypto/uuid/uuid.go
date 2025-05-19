package uuid

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	UUID_VARIANT_RESERVED_NCS_BACKWARD_COMPATIBLE = 0x00
	UUID_VARIANT_RESERVED_RFC_4122_NAMESPACE_DNS  = 0x01
	UUID_VARIANT_RESERVED_RFC_4122_NAMESPACE_URL  = 0x02
	UUID_VARIANT_RESERVED_RFC_4122_NAMESPACE_OID  = 0x03
	UUID_VARIANT_RESERVED_RFC_4122_NAMESPACE_X500 = 0x04
	UUID_VARIANT_RESERVED_FUTURE_USE              = 0x05
	UUID_VARIANT_RESERVED_FOR_NCS_COMPATIBILITY   = 0x06
	UUID_VARIANT_RESERVED_FOR_FUTURE_USE          = 0x07
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

	data7high := (u.Data[7] & 0xF0) >> 4
	data7low := u.Data[7] & 0x0F

	// Set the version in the 6th byte
	data[6] = (u.Version&0xF)<<4 | data6high&0xF

	// Copy the 7th byte of the UUID (unchanged)
	data[7] = data6low&0xF<<4 | data7high&0xF

	// Set the variant in the 8th byte
	data[8] = (u.Variant&0xF)<<4 | data7low&0xF

	// Copy the last 7 bytes of the UUID
	copy(data[9:9+7], u.Data[8:8+7])

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
func (u *UUID) Unmarshal(marshalledData []byte) (int, error) {
	if len(marshalledData) < 16 {
		return 0, fmt.Errorf("invalid UUID length: got %d bytes, want 16 bytes", len(marshalledData))
	}

	// Extract version from the 6th byte
	u.Version = (marshalledData[6] & 0xF0) >> 4

	// Extract variant from the 8th byte
	u.Variant = (marshalledData[8] & 0xF0) >> 4

	// Copy the first 6 bytes of the UUID
	copy(u.Data[0:6], marshalledData[0:6])

	u.Data[6] = (marshalledData[6]&0x0F)<<4 | (marshalledData[7]&0xF0)>>4

	u.Data[7] = (marshalledData[7]&0x0F)<<4 | (marshalledData[8] & 0x0F)

	// Copy the last 7 bytes (9-15)
	copy(u.Data[8:8+7], marshalledData[9:9+7])

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

	marshalledData, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("invalid UUID string format: %v", err)
	}

	// Use Unmarshal to populate the UUID structure
	_, err = u.Unmarshal(marshalledData)
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
