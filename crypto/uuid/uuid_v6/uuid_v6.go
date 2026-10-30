package uuid_v6

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/uuid"
)

const (
	// UUIDv6Epoch is the epoch of the UUIDv6 timestamp
	// It is the number of 100-nanosecond intervals since October 15, 1582
	UUIDv6Epoch = uint64(122192928000000000)
)

// UUIDv6 represents a UUID v6 structure (time-ordered variant of v1).
//
// UUIDv6 reorders the v1 timestamp fields to improve lexicographic ordering
// while retaining the same 60-bit timestamp, clock sequence and node fields.
type UUIDv6 struct {
	uuid.UUID

	Time uint64

	ClockSeq uint16

	NodeID [6]byte
}

// Marshal converts the UUIDv6 structure to a 16-byte array
//
// Returns:
//   - A byte slice containing the UUID's 16 bytes
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv6) Marshal() ([]byte, error) {
	// Create a 15-byte array to hold the UUID data
	var data [15]byte

	u.UUID.Data = [15]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// UUIDv6 timestamp layout (60 bits):
	// time_high (32 bits) | time_mid (16 bits) | time_low (12 bits)
	timeHigh := uint32((u.Time >> 28) & 0xFFFFFFFF)
	timeMid := uint16((u.Time >> 12) & 0xFFFF)
	timeLow := uint16(u.Time & 0x0FFF)

	// Place time fields in reordered positions
	binary.BigEndian.PutUint32(data[0:4], timeHigh)
	binary.BigEndian.PutUint16(data[4:6], timeMid)

	// Encode 12-bit timeLow across data[6] and data[7] high nibble
	data[6] = byte((timeLow >> 4) & 0xFF)
	data[7] = byte(timeLow&0x0F)<<4 | byte((u.ClockSeq&0x0F00)>>8)

	data[8] = byte(u.ClockSeq & 0xFF)

	// Copy node ID to the remaining bytes
	copy(data[9:15], u.NodeID[:])

	// Set the UUID version (6) and variant. The variant nibble carries the
	// RFC 9562 variant (0b10) followed by the top 2 bits of the 14-bit
	// clock sequence.
	u.UUID.Version = 6
	u.UUID.Variant = 0x8 | byte((u.ClockSeq>>12)&0x3)
	u.UUID.Data = data

	// Use the UUID's Marshal method to get the final 16-byte array
	return u.UUID.Marshal()
}

// Unmarshal converts a 16-byte array into a UUIDv6 structure
//
// Returns:
//   - The number of bytes read
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv6) Unmarshal(marshalledData []byte) (int, error) {
	if len(marshalledData) < 16 {
		return 0, fmt.Errorf("invalid UUID length: got %d bytes, want 16 bytes", len(marshalledData))
	}

	// First unmarshal into the generic UUID
	n, err := u.UUID.Unmarshal(marshalledData)
	if err != nil {
		return 0, err
	}

	// Check if this is a version 6 UUID
	if u.UUID.Version != 6 {
		return 0, fmt.Errorf("invalid UUID version: got %d, want 6", u.UUID.Version)
	}

	// Extract time fields from the UUID data (reordered layout)
	timeHigh := binary.BigEndian.Uint32(u.UUID.Data[0:4])
	timeMid := binary.BigEndian.Uint16(u.UUID.Data[4:6])
	// Extract 12-bit timeLow
	timeLow := uint16(u.UUID.Data[6])<<4 | (uint16(u.UUID.Data[7]>>4) & 0x0F)

	// Extract the 14-bit clock sequence; its top 2 bits follow the variant bits
	u.ClockSeq = uint16(u.UUID.Variant&0x3)<<12 | uint16(u.UUID.Data[7]&0x0F)<<8 | uint16(u.UUID.Data[8])

	// Reconstruct the time field
	u.Time = (uint64(timeHigh) << 28) | (uint64(timeMid) << 12) | uint64(timeLow)

	// Copy node ID
	copy(u.NodeID[:], u.UUID.Data[9:15])

	return n, nil
}

// FromString parses a UUID string in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
// and returns a UUIDv6 structure.
//
// Parameters:
//   - uuidStr: A string containing the UUID in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
//
// Returns:
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv6) FromString(uuidStr string) error {

	uuidStr = strings.ReplaceAll(uuidStr, "-", "")

	if len(uuidStr) != 32 {
		return fmt.Errorf("invalid UUID length: got %d characters, want 32", len(uuidStr))
	}

	uuidStr = strings.ToLower(uuidStr)

	uuidBytes, err := hex.DecodeString(uuidStr)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	_, err = u.Unmarshal(uuidBytes)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	return nil
}

// GetTime returns the time of the UUIDv6 structure
//
// Returns:
//   - The time of the UUIDv6 structure
func (u *UUIDv6) GetTime() time.Time {
	// UUID v6 timestamp is 100-nanosecond intervals since October 15, 1582
	// We need to convert to Unix time (January 1, 1970)

	// Extract the timestamp from the UUID
	timestamp := uint64(u.Time)

	// Convert to Unix timestamp (in nanoseconds)
	unixNs := int64((timestamp - UUIDv6Epoch) * 100)

	// Create time from Unix nanoseconds
	return time.Unix(0, unixNs)
}

// GetNodeID returns the node ID of the UUIDv6 structure
//
// Returns:
//   - The node ID of the UUIDv6 structure
func (u *UUIDv6) GetNodeID() []byte {
	return u.NodeID[:]
}

// GetClockSequence returns the clock sequence of the UUIDv6 structure
//
// Returns:
//   - The clock sequence of the UUIDv6 structure
func (u *UUIDv6) GetClockSequence() uint16 {
	// The clock sequence is stored in the ClockSeqAndNodeID field
	// The high bits are in the first byte, and the low bits are in the second byte
	return u.ClockSeq
}

// SetTime sets the time field of the UUIDv6 structure
//
// Parameters:
//   - t: The time to set
func (u *UUIDv6) SetTime(t time.Time) {
	// Convert Unix time to UUID v6 timestamp (100-nanosecond intervals since October 15, 1582)
	unixNs := t.UnixNano()
	timestamp := uint64(unixNs/100) + UUIDv6Epoch
	u.Time = timestamp

	// Update the UUID data fields related to time
	u.UUID.Version = 6
}

// SetNodeID sets the node ID field of the UUIDv6 structure
//
// Parameters:
//   - nodeID: A byte slice containing the node ID (6 bytes)
//
// Returns:
//   - An error if the node ID is invalid
func (u *UUIDv6) SetNodeID(nodeID []byte) error {
	if len(nodeID) != 6 {
		return fmt.Errorf("invalid node ID length: got %d bytes, want 6", len(nodeID))
	}
	copy(u.NodeID[:], nodeID)
	return nil
}

// SetClockSequence sets the clock sequence field of the UUIDv6 structure
//
// Parameters:
//   - clockSeq: The clock sequence to set
func (u *UUIDv6) SetClockSequence(clockSeq uint16) {
	u.ClockSeq = clockSeq
}

// String returns the string representation of the UUIDv6 structure
//
// Returns:
//   - The string representation of the UUIDv6 structure
func (u *UUIDv6) String() string {
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
