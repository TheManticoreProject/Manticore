package types

import (
	"encoding/binary"
	"fmt"
)

// LOCKING_ANDX_RANGE64
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/b5c6eae7-976b-4444-b52e-c76c68c861ad
type LOCKING_ANDX_RANGE64 struct {
	// PID (2 bytes): The PID of the process requesting the locking change.
	PID USHORT

	// Pad (2 bytes): This field pads the structure to DWORD alignment and MUST be zero (0x0000).
	Pad USHORT

	// OffsetInBytesHigh (4 bytes): The 32-bit unsigned integer value that is the high 32 bits of a
	// 64-bit offset into the file at which the locking change MUST begin.
	ByteOffsetHigh ULONG

	// OffsetInBytesLow (4 bytes): The 32-bit unsigned integer value that is the low 32 bits of a
	// 64-bit offset into the file at which the locking change MUST begin.
	ByteOffsetLow ULONG

	// LengthInBytesHigh (4 bytes): The 32-bit unsigned integer value that is the high 32 bits of a
	// 64-bit value specifying the number of bytes that MUST be locked or unlocked.
	LengthInBytesHigh ULONG

	// LengthInBytesLow (4 bytes): The 32-bit unsigned integer value that is the low 32 bits of a
	// 64-bit value specifying the number of bytes that MUST be locked or unlocked.
	LengthInBytesLow ULONG
}

// Marshal marshals the LOCKING_ANDX_RANGE64 structure into a byte array
//
// Returns:
// - A byte array representing the LOCKING_ANDX_RANGE64 structure
// - An error if the marshaling fails
func (l *LOCKING_ANDX_RANGE64) Marshal() ([]byte, error) {
	result := make([]byte, 20) // 2 + 2 + 4 + 4 + 4 + 4 bytes

	// Marshal PID (2 bytes)
	binary.LittleEndian.PutUint16(result[0:2], uint16(l.PID))

	// Marshal Pad (2 bytes)
	binary.LittleEndian.PutUint16(result[2:4], uint16(l.Pad))

	// Marshal ByteOffsetHigh (4 bytes)
	binary.LittleEndian.PutUint32(result[4:8], uint32(l.ByteOffsetHigh))

	// Marshal ByteOffsetLow (4 bytes)
	binary.LittleEndian.PutUint32(result[8:12], uint32(l.ByteOffsetLow))

	// Marshal LengthInBytesHigh (4 bytes)
	binary.LittleEndian.PutUint32(result[12:16], uint32(l.LengthInBytesHigh))

	// Marshal LengthInBytesLow (4 bytes)
	binary.LittleEndian.PutUint32(result[16:20], uint32(l.LengthInBytesLow))

	return result, nil
}

// Unmarshal unmarshals a byte array into the LOCKING_ANDX_RANGE64 structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
// - An error if the unmarshaling fails
func (l *LOCKING_ANDX_RANGE64) Unmarshal(data []byte) (int, error) {
	if len(data) < 20 {
		return 0, fmt.Errorf("data too short for LOCKING_ANDX_RANGE64")
	}

	// Unmarshal PID (2 bytes)
	l.PID = USHORT(binary.LittleEndian.Uint16(data[0:2]))

	// Unmarshal Pad (2 bytes)
	l.Pad = USHORT(binary.LittleEndian.Uint16(data[2:4]))

	// Unmarshal ByteOffsetHigh (4 bytes)
	l.ByteOffsetHigh = ULONG(binary.LittleEndian.Uint32(data[4:8]))

	// Unmarshal ByteOffsetLow (4 bytes)
	l.ByteOffsetLow = ULONG(binary.LittleEndian.Uint32(data[8:12]))

	// Unmarshal LengthInBytesHigh (4 bytes)
	l.LengthInBytesHigh = ULONG(binary.LittleEndian.Uint32(data[12:16]))

	// Unmarshal LengthInBytesLow (4 bytes)
	l.LengthInBytesLow = ULONG(binary.LittleEndian.Uint32(data[16:20]))

	return 20, nil
}
