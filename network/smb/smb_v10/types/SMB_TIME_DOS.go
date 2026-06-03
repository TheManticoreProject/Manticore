package types

import (
	"encoding/binary"
	"fmt"
)

// SMB_TIME_DOS is the 2-byte MS-DOS time representation used by legacy SMB
// commands (e.g. SMB_COM_SET_INFORMATION2). It packs the time of day into a
// little-endian 16-bit value, analogous to the 2-byte SMB_DATE type.
//
// The 8-byte FILETIME-backed types.SMB_TIME alias is unsuitable for these
// fields because the wire format mandates exactly 2 bytes per SMB_TIME.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3e6f3a13-6a40-4f76-af70-bb514554ea5b
type SMB_TIME_DOS struct {
	// Hours: bits 11-15 (0xF800) - Values 0-23
	Hours uint8
	// Minutes: bits 5-10 (0x07E0) - Values 0-59
	Minutes uint8
	// TwoSeconds: bits 0-4 (0x001F) - Values 0-29, each unit is 2 seconds
	TwoSeconds uint8
}

// NewSMB_TIME_DOS creates a new SMB_TIME_DOS structure
//
// Returns:
// - A pointer to the new SMB_TIME_DOS structure
func NewSMB_TIME_DOS() *SMB_TIME_DOS {
	return &SMB_TIME_DOS{}
}

// NewSMB_TIME_DOSFromTime creates a new SMB_TIME_DOS structure from a time of day
//
// Parameters:
// - hours: The hours of the time (0-23)
// - minutes: The minutes of the time (0-59)
// - seconds: The seconds of the time (0-59); stored with 2-second resolution
//
// Returns:
// - The new SMB_TIME_DOS structure
func NewSMB_TIME_DOSFromTime(hours int, minutes int, seconds int) *SMB_TIME_DOS {
	return &SMB_TIME_DOS{
		Hours:      uint8(hours),
		Minutes:    uint8(minutes),
		TwoSeconds: uint8(seconds / 2),
	}
}

// Marshal marshals the SMB_TIME_DOS structure
//
// Returns:
// - A byte array representing the SMB_TIME_DOS structure
// - An error if the marshaling fails
func (t *SMB_TIME_DOS) Marshal() ([]byte, error) {
	marshalledData := [2]byte{}

	// Encode the time according to the SMB_TIME format:
	// - Hours: bits 11-15 (0xF800)
	// - Minutes: bits 5-10 (0x07E0)
	// - TwoSeconds: bits 0-4 (0x001F)
	valueHours := uint16(t.Hours) << 11
	valueMinutes := uint16(t.Minutes) << 5
	valueTwoSeconds := uint16(t.TwoSeconds)

	value := valueHours | valueMinutes | valueTwoSeconds

	// Convert to little-endian byte order
	binary.LittleEndian.PutUint16(marshalledData[:], value)

	return marshalledData[:], nil
}

// Unmarshal unmarshals the SMB_TIME_DOS structure
//
// Returns:
// - The number of bytes unmarshalled
// - An error if the unmarshaling fails
func (t *SMB_TIME_DOS) Unmarshal(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("data too short for SMB_TIME_DOS")
	}

	value := binary.LittleEndian.Uint16(data[:2])

	// Extract the hours value (bits 11-15)
	t.Hours = uint8((value & 0xF800) >> 11)

	// Extract the minutes value (bits 5-10)
	t.Minutes = uint8((value & 0x07E0) >> 5)

	// Extract the two-seconds value (bits 0-4)
	t.TwoSeconds = uint8(value & 0x001F)

	return 2, nil
}
