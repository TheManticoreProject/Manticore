package types

import (
	"encoding/binary"
	"fmt"
)

// SMB_DATE
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/31b65222-4171-49b4-aeed-7d3f38ecf68b
type SMB_DATE struct {
	Year  uint16
	Month uint8
	Day   uint8
}

// NewSMB_DATE creates a new SMB_DATE structure
//
// Returns:
// - A pointer to the new SMB_DATE structure
func NewSMB_DATE() *SMB_DATE {
	return &SMB_DATE{}
}

// NewSMB_DATEFromDate creates a new SMB_DATE structure from a date
//
// Parameters:
// - year: The year of the date
// - month: The month of the date
// - day: The day of the date
//
// Returns:
// - The new SMB_DATE structure
func NewSMB_DATEFromDate(year int, month int, day int) *SMB_DATE {
	return &SMB_DATE{
		Year:  uint16(year),
		Month: uint8(month),
		Day:   uint8(day),
	}
}

// Marshal marshals the SMB_DATE structure
//
// Returns:
// - A byte array representing the SMB_DATE structure
// - An error if the marshaling fails
func (d *SMB_DATE) Marshal() ([]byte, error) {
	marshalledData := [2]byte{}

	value := uint16(0)

	// Encode the date according to the SMB_DATE format:
	// - Year: bits 9-15 (0xFE00) - Add 1980 to get actual year
	// - Month: bits 5-8 (0x01E0) - Values 1-12
	// - Day: bits 0-4 (0x001F) - Values 1-31

	// Calculate the year value (subtract 1980 and shift to bits 9-15)
	valueYear := (d.Year - 1980) << 9

	// Calculate the month value (shift to bits 5-8)
	valueMonth := uint16(d.Month) << 5

	// Day value uses bits 0-4
	valueDay := uint16(d.Day)

	// Combine all values into a 16-bit value
	value = valueYear | valueMonth | valueDay

	// Convert to little-endian byte order
	binary.LittleEndian.PutUint16(marshalledData[:], value)

	return marshalledData[:], nil
}

// Unmarshal unmarshals the SMB_DATE structure
//
// Returns:
// - The number of bytes unmarshalled
// - An error if the unmarshaling fails
func (d *SMB_DATE) Unmarshal(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("data too short for SMB_DATE")
	}

	value := binary.LittleEndian.Uint16(data[:2])

	// Extract the year value (bits 9-15)
	yearValue := (value & 0xFE00) >> 9

	// Extract the month value (bits 5-8)
	monthValue := (value & 0x01E0) >> 5

	// Extract the day value (bits 0-4)
	dayValue := value & 0x001F

	d.Year = uint16(yearValue + 1980)
	d.Month = uint8(monthValue)
	d.Day = uint8(dayValue)

	return 2, nil
}
