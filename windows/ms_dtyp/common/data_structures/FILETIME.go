package data_structures

import (
	"encoding/binary"
	"errors"
	"time"
)

// Thursday, January 1, 1970 1:00:00 AM GMT+01:00 in 100-nanosecond intervals
const UnixTimestampIn100NsIntervals int64 = 116444736000000000

// The FILETIME structure is a 64-bit value that represents the number of 100-nanosecond intervals
// that have elapsed since January 1, 1601, Coordinated Universal Time (UTC).
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/2c57429b-fdd4-488f-b5fc-9e4cf020fcdf
type FILETIME struct {
	// dwLowDateTime: A 32-bit unsigned integer that contains the low-order bits of the file time.
	DwLowDateTime uint32
	// dwHighDateTime: A 32-bit unsigned integer that contains the high-order bits of the file time.
	DwHighDateTime uint32
}

// NewFILETIMEFromTime creates a new FILETIME structure from a time.Time value.
//
// Parameters:
// - t: The time.Time value to create the FILETIME structure from
//
// Returns:
// - A pointer to the new FILETIME structure
func NewFILETIMEFromTime(t time.Time) *FILETIME {
	value := (t.UnixNano() / 100) + UnixTimestampIn100NsIntervals
	return &FILETIME{
		DwLowDateTime:  uint32(value & 0xFFFFFFFF),
		DwHighDateTime: uint32((value >> 32) & 0xFFFFFFFF),
	}
}

// ToInt64 returns the int64 representation of the FILETIME structure.
//
// Returns:
// - The int64 representation of the FILETIME structure
func (ft *FILETIME) ToInt64() int64 {
	return (int64(ft.DwHighDateTime) & 0xFFFFFFFF << 32) | (int64(ft.DwLowDateTime) & 0xFFFFFFFF)
}

// GetTime returns the time represented by the FILETIME structure.
//
// Returns:
// - The time represented by the FILETIME structure
func (ft *FILETIME) GetTime() time.Time {
	delta := int64(ft.ToInt64()-UnixTimestampIn100NsIntervals) * 100
	return time.Unix(0, delta)
}

// Unmarshal deserializes a byte slice into the FILETIME structure.
//
// Parameters:
// - data: A byte slice to be deserialized into the FILETIME structure
func (ft *FILETIME) Unmarshal(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, errors.New("data is too short to unmarshal into FILETIME")
	}

	ft.DwLowDateTime = binary.LittleEndian.Uint32(data[0:4])

	ft.DwHighDateTime = binary.LittleEndian.Uint32(data[4:8])

	return 8, nil
}

// Marshal serializes the FILETIME structure into a byte slice.
// It converts the FILETIME structure into its binary representation
// according to the SMB protocol format.
//
// Returns:
// - A byte slice containing the marshalled FILETIME structure
func (ft *FILETIME) Marshal() ([]byte, error) {
	marshalledData := []byte{}

	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, ft.DwLowDateTime)
	marshalledData = append(marshalledData, buf4...)

	buf4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, ft.DwHighDateTime)
	marshalledData = append(marshalledData, buf4...)

	return marshalledData, nil
}

// GetUnixTimestamp returns the Unix timestamp represented by the FILETIME structure.
//
// Returns:
// - The Unix timestamp as an int64
func (ft *FILETIME) GetUnixTimestamp() int64 {
	// Convert Windows file time (100-nanosecond intervals since January 1, 1601)
	// to Unix time (seconds since January 1, 1970)
	return ft.GetTime().Unix()
}

// GetTimeString returns the string representation of the FILETIME structure in UTC.
//
// Returns:
// - The string representation of the FILETIME structure in UTC
func (ft *FILETIME) GetTimeString() string {
	return ft.GetTime().UTC().Format("2006-01-02 15:04:05.00000")
}

// String returns the string representation of the FILETIME structure in UTC.
//
// Returns:
// - The string representation of the FILETIME structure in UTC
func (ft *FILETIME) String() string {
	return ft.GetTimeString()
}
