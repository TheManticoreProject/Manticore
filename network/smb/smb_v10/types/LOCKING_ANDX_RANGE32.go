package types

import (
	"encoding/binary"
	"fmt"
)

// LOCKING_ANDX_RANGE32
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/b5c6eae7-976b-4444-b52e-c76c68c861ad
type LOCKING_ANDX_RANGE32 struct {
	// PID (2 bytes): The PID of the process requesting the locking change.
	PID USHORT

	// ByteOffset (4 bytes): The 32-bit unsigned integer value that is the offset into the file at which
	// the locking change MUST begin.
	ByteOffset ULONG

	// LengthInBytes (4 bytes): The 32-bit unsigned integer value that is the number of bytes, beginning
	// at OffsetInBytes, that MUST be locked or unlocked.
	LengthInBytes ULONG
}

// Marshal marshals the LOCKING_ANDX_RANGE32 structure into a byte array
//
// Returns:
// - A byte array representing the LOCKING_ANDX_RANGE32 structure
// - An error if the marshaling fails
func (l *LOCKING_ANDX_RANGE32) Marshal() ([]byte, error) {
	result := make([]byte, 10) // 2 + 4 + 4 bytes

	// Marshal PID (2 bytes)
	binary.LittleEndian.PutUint16(result[0:2], uint16(l.PID))

	// Marshal ByteOffset (4 bytes)
	binary.LittleEndian.PutUint32(result[2:6], uint32(l.ByteOffset))

	// Marshal LengthInBytes (4 bytes)
	binary.LittleEndian.PutUint32(result[6:10], uint32(l.LengthInBytes))

	return result, nil
}

// Unmarshal unmarshals a byte array into the LOCKING_ANDX_RANGE32 structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
// - An error if the unmarshaling fails
func (l *LOCKING_ANDX_RANGE32) Unmarshal(data []byte) (int, error) {
	if len(data) < 10 {
		return 0, fmt.Errorf("data too short for LOCKING_ANDX_RANGE32")
	}

	// Unmarshal PID (2 bytes)
	l.PID = USHORT(binary.LittleEndian.Uint16(data[0:2]))

	// Unmarshal ByteOffset (4 bytes)
	l.ByteOffset = ULONG(binary.LittleEndian.Uint32(data[2:6]))

	// Unmarshal LengthInBytes (4 bytes)
	l.LengthInBytes = ULONG(binary.LittleEndian.Uint32(data[6:10]))

	return 10, nil
}
