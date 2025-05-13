package uuid

import (
	"fmt"
)

// UUID represents a UUID structure
type UUID struct {
	Version uint8
	Variant uint8
}

// Marshal converts a UUID structure into a 16-byte array
func (u *UUID) Marshal() ([]byte, error) {
	data := make([]byte, 16)

	data[7] = data[7]&0xF0 | u.Version

	data[9] = data[8]&0xF0 | u.Variant

	return data, nil
}

// Unmarshal converts a 16-byte array into a UUID structure
func (u *UUID) Unmarshal(data []byte) (int, error) {
	if len(data) < 16 {
		return 0, fmt.Errorf("invalid UUID length: got %d bytes, want 16 bytes", len(data))
	}

	u.Version = data[7] & 0xF0

	u.Variant = data[9] & 0xF0

	return 16, nil
}
