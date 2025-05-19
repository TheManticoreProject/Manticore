package uuid_v2

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/uuid"
)

const (
	// UUIDv2Epoch is the epoch of the UUIDv2 timestamp
	// It is the number of 100-nanosecond intervals since October 15, 1582
	UUIDv2Epoch = uint64(122192928000000000)
)

// UUIDv2 represents a UUID v1 structure
//
// UUIDv2 is a structure that represents a UUID v1.
// It contains a UUID, a time, a clock sequence, and a node ID.
type UUIDv2 struct {
	uuid.UUID

	Time uint64

	Clock uint8

	LocalDomain uint8

	NodeID [6]byte
}

// Marshal converts the UUIDv2 structure to a 16-byte array
//
// Returns:
//   - A byte slice containing the UUID's 16 bytes
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv2) Marshal() ([]byte, error) {
	// Create a 15-byte array to hold the UUID data
	var data [15]byte

	u.UUID.Data = [15]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// Set time fields (the first 6 bytes of information)
	timeLow := uint32(u.Time & 0x00000000FFFFFFFF)
	binary.BigEndian.PutUint32(data[0:4], timeLow)

	timeMid := uint16((u.Time & 0x0000FFFF00000000) >> 32)
	binary.BigEndian.PutUint16(data[4:6], timeMid)

	// In this part the encoding is a bit tricky
	// If timeHigh is 0xAAA and clockSeq is 0xBBB
	// We need to encode it as 0xAAABBB
	timeHigh := uint16((u.Time & 0x0FFF000000000000) >> 48)
	data[6] = byte((timeHigh >> 4) & 0xFF)

	data[7] = byte(timeHigh&0x0F)<<4 | byte(u.Clock&0x0F)

	data[8] = u.LocalDomain

	// Copy node ID to the remaining bytes
	copy(data[9:15], u.NodeID[:])

	// Set the UUID version and variant
	u.UUID.Version = 2
	u.UUID.Data = data

	// Use the UUID's Marshal method to get the final 16-byte array
	return u.UUID.Marshal()
}

// Unmarshal converts a 16-byte array into a UUIDv2 structure
//
// Returns:
//   - The number of bytes read
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv2) Unmarshal(marshalledData []byte) (int, error) {
	if len(marshalledData) < 16 {
		return 0, fmt.Errorf("invalid UUID length: got %d bytes, want 16 bytes", len(marshalledData))
	}

	// First unmarshal into the generic UUID
	n, err := u.UUID.Unmarshal(marshalledData)
	if err != nil {
		return 0, err
	}

	// Check if this is a version 2 UUID
	if u.UUID.Version != 2 {
		return 0, fmt.Errorf("invalid UUID version: got %d, want 2", u.UUID.Version)
	}

	// Extract time fields from the UUID data
	timeLow := binary.BigEndian.Uint32(u.UUID.Data[0:4])

	timeMid := binary.BigEndian.Uint16(u.UUID.Data[4:6])

	// Extract time high
	timeHigh := uint16(u.UUID.Data[6])<<4 | uint16(u.UUID.Data[7]>>4)&0xF

	// Extract clock sequence
	u.Clock = u.UUID.Data[7] & 0x0F

	u.LocalDomain = u.UUID.Data[8]

	// Reconstruct the time field
	u.Time = uint64(timeHigh)<<48 | uint64(timeMid)<<32 | uint64(timeLow)

	// Copy node ID
	copy(u.NodeID[:], u.UUID.Data[9:15])

	return n, nil
}

// FromString parses a UUID string in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
// and returns a UUIDv2 structure.
//
// Parameters:
//   - uuidStr: A string containing the UUID in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
//
// Returns:
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv2) FromString(uuidStr string) error {

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

// FromBytes creates a UUIDv2 from a 16-byte array
//
// Parameters:
//   - data: A byte slice containing the UUID's 16 bytes
//
// Returns:
//   - An error if the UUID is invalid or the conversion fails
func (u *UUIDv2) FromBytes(data []byte) error {
	if len(data) != 16 {
		return fmt.Errorf("invalid UUID length: got %d bytes, want 16", len(data))
	}
	_, err := u.Unmarshal(data)
	return err
}

// GetTime returns the time of the UUIDv2 structure
//
// Returns:
//   - The time of the UUIDv2 structure
func (u *UUIDv2) GetTime() time.Time {
	// UUID v1 timestamp is 100-nanosecond intervals since October 15, 1582
	// We need to convert to Unix time (January 1, 1970)

	// Extract the timestamp from the UUID
	timestamp := uint64(u.Time)

	// Convert to Unix timestamp (in nanoseconds)
	unixNs := int64((timestamp - UUIDv2Epoch) * 100)

	// Create time from Unix nanoseconds
	return time.Unix(0, unixNs)
}

// GetNodeID returns the node ID of the UUIDv2 structure
//
// Returns:
//   - The node ID of the UUIDv2 structure
func (u *UUIDv2) GetNodeID() []byte {
	return u.NodeID[:]
}

// GetClockuence returns the clock sequence of the UUIDv2 structure
//
// Returns:
//   - The clock sequence of the UUIDv2 structure
func (u *UUIDv2) GetClock() uint8 {
	return u.Clock
}

// GetLocalDomain returns the local domain of the UUIDv2 structure
//
// Returns:
//   - The local domain of the UUIDv2 structure
func (u *UUIDv2) GetLocalDomain() uint8 {
	return u.LocalDomain
}

// SetTime sets the time field of the UUIDv2 structure
//
// Parameters:
//   - t: The time to set
func (u *UUIDv2) SetTime(t time.Time) {
	// Convert Unix time to UUID v1 timestamp (100-nanosecond intervals since October 15, 1582)
	unixNs := t.UnixNano()
	timestamp := uint64(unixNs/100) + UUIDv2Epoch
	u.Time = timestamp

	// Update the UUID data fields related to time
	u.UUID.Version = 1
}

// SetNodeID sets the node ID field of the UUIDv2 structure
//
// Parameters:
//   - nodeID: A byte slice containing the node ID (6 bytes)
//
// Returns:
//   - An error if the node ID is invalid
func (u *UUIDv2) SetNodeID(nodeID []byte) error {
	if len(nodeID) != 6 {
		return fmt.Errorf("invalid node ID length: got %d bytes, want 6", len(nodeID))
	}
	copy(u.NodeID[:], nodeID)
	return nil
}

// SetClockuence sets the clock sequence field of the UUIDv2 structure
//
// Parameters:
//   - clockSeq: The clock sequence to set
func (u *UUIDv2) SetClock(clock uint8) {
	u.Clock = clock
}

// SetLocalDomain sets the local domain field of the UUIDv2 structure
//
// Parameters:
//   - localDomain: The local domain to set
func (u *UUIDv2) SetLocalDomain(localDomain uint8) {
	u.LocalDomain = localDomain
}

// String returns the string representation of the UUIDv2 structure
//
// Returns:
//   - The string representation of the UUIDv2 structure
func (u *UUIDv2) String() string {
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
