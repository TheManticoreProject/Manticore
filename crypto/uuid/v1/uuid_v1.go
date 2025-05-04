package uuid_v1

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	UUIDv1_VARIANT_RESERVED_NCS_BACKWARD_COMPATIBLE = 0x00
	UUIDv1_VARIANT_RESERVED_RFC_4122_NAMESPACE_DNS  = 0x01
	UUIDv1_VARIANT_RESERVED_RFC_4122_NAMESPACE_URL  = 0x02
	UUIDv1_VARIANT_RESERVED_RFC_4122_NAMESPACE_OID  = 0x03
	UUIDv1_VARIANT_RESERVED_RFC_4122_NAMESPACE_X500 = 0x04
	UUIDv1_VARIANT_RESERVED_FUTURE_USE              = 0x05
	UUIDv1_VARIANT_RESERVED_FOR_NCS_COMPATIBILITY   = 0x06
	UUIDv1_VARIANT_RESERVED_FOR_FUTURE_USE          = 0x07

	UUIDv1Epoch = uint64(122192928000000000)
)

// UUIDv1 represents a UUID v1 structure
type UUIDv1 struct {
	Time     uint64
	Version  uint8
	Variant  uint8
	ClockSeq uint16
	NodeID   [6]byte
}

// Unmarshal converts a 16-byte array into a UUIDv1 structure
func (u *UUIDv1) Unmarshal(marshalledData []byte) (int, error) {
	if len(marshalledData) < 16 {
		return 0, fmt.Errorf("invalid UUID length: got %d bytes, want 16 bytes", len(marshalledData))
	}

	// Extract the variant
	// The variant is stored in the most significant bits of the 8th octet
	u.Variant = marshalledData[8] >> 4

	// Reconstruct the time field
	u.Time = binary.LittleEndian.Uint64(marshalledData[0:8])

	fmt.Printf("marshalledData: %x\n", marshalledData)

	// Extract version
	u.Version = uint8((marshalledData[8] >> 4) & 0x0F)
	if u.Version != 1 {
		return 0, fmt.Errorf("invalid UUID version: got %d, want 1", u.Version)
	}

	// Extract clock sequence and node ID
	u.ClockSeq = binary.LittleEndian.Uint16(marshalledData[8:10])

	// Copy node ID
	copy(u.NodeID[:], marshalledData[10:16])

	return 16, nil
}

// Decode parses a UUID string in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
// and returns a UUIDv1 structure.
func (u *UUIDv1) FromString(uuidStr string) error {
	if len(uuidStr) != 36 {
		return fmt.Errorf("invalid UUID length: got %d characters, want 36", len(uuidStr))
	}

	parts := strings.Split(uuidStr, "-")

	if len(parts) != 5 {
		return fmt.Errorf("invalid UUID format: expected 5 parts separated by hyphens")
	}
	bytesStream := []byte{}
	for k, part := range parts {
		decoded, err := hex.DecodeString(part)
		if err != nil {
			return fmt.Errorf("invalid UUID format for part %d: %v", k, err)
		}
		bytesStream = append(bytesStream, decoded...)
	}

	_, err := u.Unmarshal(bytesStream)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	return nil
}

// FromBytes creates a UUIDv1 from a 16-byte array
func (u *UUIDv1) FromBytes(data []byte) error {
	if len(data) != 16 {
		return fmt.Errorf("invalid UUID length: got %d bytes, want 16", len(data))
	}
	_, err := u.Unmarshal(data)
	return err
}

// Marshal converts the UUIDv1 structure to a 16-byte array
func (u *UUIDv1) Marshal() ([]byte, error) {
	data := []byte{}

	timeHigh := (u.Time & 0xFFFF000000000000) >> 48
	timeMid := (u.Time & 0x0000FFFF0000) >> 16
	timeLow := u.Time & 0x000000000000FFFF

	clockSeqHigh := (u.ClockSeq & 0xFF00) >> 8
	clockSeqLow := u.ClockSeq & 0x00FF

	copy(data, []byte{
		byte(timeHigh),
		byte(timeMid),
		byte(timeLow),
		byte(clockSeqHigh),
		byte(clockSeqLow),
	})

	copy(data, u.NodeID[:])

	return data, nil
}

// GetTime returns the time of the UUIDv1 structure
func (u *UUIDv1) GetTime() time.Time {
	// UUID v1 timestamp is 100-nanosecond intervals since October 15, 1582
	// We need to convert to Unix time (January 1, 1970)

	// Extract the timestamp from the UUID
	timestamp := uint64(u.Time)

	// Number of 100-nanosecond intervals between UUID epoch and Unix epoch
	// UUID epoch: October 15, 1582
	// Unix epoch: January 1, 1970
	epochDiff := uint64(122192928000000000)

	// Convert to Unix timestamp (in nanoseconds)
	unixNs := int64((timestamp - epochDiff) * 100)

	// Create time from Unix nanoseconds
	return time.Unix(0, unixNs)
}

// GetNodeID returns the node ID of the UUIDv1 structure
func (u *UUIDv1) GetNodeID() []byte {
	return u.NodeID[:]
}

// GetClockSequence returns the clock sequence of the UUIDv1 structure
func (u *UUIDv1) GetClockSequence() uint16 {
	// The clock sequence is stored in the ClockSeqAndNodeID field
	// The high bits are in the first byte, and the low bits are in the second byte
	return u.ClockSeq
}

// String returns the string representation of the UUIDv1 structure
func (u *UUIDv1) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		u.Time,
		u.Version,
		u.Variant,
		u.ClockSeq,
		u.NodeID,
	)
}
