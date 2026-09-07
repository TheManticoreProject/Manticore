package utils

import (
	"encoding/binary"
	"fmt"
	"time"
)

// DateTime represents a point in time with additional precision.
//
// Fields:
// - Time: A time.Time object representing the point in time.
// - Ticks: A uint64 value representing the number of 100-nanosecond intervals that have elapsed since 1601-01-01 00:00:00 UTC.
//
// Methods:
// - NewDateTime: Initializes a new DateTime instance based on the provided ticks or the current time if ticks is 0.
// - ToUniversalTime: Converts the DateTime instance to a time.Time object in UTC.
//
// Note:
// The Ticks field provides additional precision for representing time, which is useful for certain applications that require high-resolution timestamps.
// The Time field is a standard time.Time object that can be used with Go's time package functions.
type DateTime struct {
	time  time.Time
	ticks uint64
}

const ticksBetween1601AndUnix uint64 = 116444736000000000 // (Unix(1970)-1601) in 100ns ticks

// ticksPerSecond is the number of 100-nanosecond intervals in one second.
const ticksPerSecond uint64 = 10000000

// NewDateTimeFromTicks initializes a new DateTime instance from a number of ticks.
//
// Parameters:
//   - ticks: A uint64 value representing the number of 100-nanosecond intervals that have elapsed since 1601-01-01 00:00:00 UTC.
//     If ticks is 0, the function sets the current time and calculates ticks from 1601 to now.
//
// Returns:
// - A DateTime object initialized with the provided ticks or the current time if ticks is 0.
//
// Note:
// The function calculates the number of nanoseconds between 1601-01-01 and the UNIX epoch (1970-01-01) to convert ticks to a time.Time object.
// If ticks is 0, the function sets the current time and calculates the ticks from 1601 to the current time.
// Otherwise, it sets the time based on the provided ticks.
func NewDateTimeFromTicks(ticks uint64) DateTime {
	dt := DateTime{}

	if ticks == 0 {
		// Capture current time without truncation to avoid returning a value
		// that can be before a caller-captured timestamp taken immediately
		// before invoking this function (due to 100ns truncation flooring).
		now := time.Now().UTC()
		epoch1601 := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
		// Use signed arithmetic to correctly support times before 1970.
		nsSince1601 := now.UnixNano() - epoch1601.UnixNano()
		dt.time = now
		dt.ticks = uint64(nsSince1601 / 100)
	} else {
		dt.SetTicks(ticks)
	}

	return dt
}

// NewDateTimeFromTime initializes a new DateTime instance from a time.Time object.
//
// Parameters:
// - t: A time.Time value to create the DateTime instance from.
//
// Returns:
// - A DateTime object initialized with the provided time.Time value.
func NewDateTimeFromTime(t time.Time) DateTime {
	dt := DateTime{}
	dt.SetTime(t)
	return dt
}

// ToUniversalTime converts the DateTime instance to a time.Time object in UTC.
//
// Returns:
// - A time.Time object representing the DateTime instance in Coordinated Universal Time (UTC).
//
// Note:
// This function ensures that the time is represented in the UTC time zone, regardless of the original time zone of the DateTime instance.
func (dt DateTime) ToUniversalTime() time.Time {
	return dt.time.UTC()
}

// SetTicks sets the number of 100-nanosecond intervals (ticks) for the DateTime instance.
//
// Parameters:
// - ticks: A uint64 value representing the number of 100-nanosecond intervals since 1601-01-01 00:00:00 UTC.
//
// Note:
// This function updates both the Ticks field and recalculates the corresponding Time field based on the provided ticks.
func (dt *DateTime) SetTicks(ticks uint64) {
	dt.ticks = ticks

	// The epoch offset is applied in whole seconds with the sub-second remainder
	// carried separately. Scaling the whole tick count to nanoseconds in one
	// multiplication exhausts the int64 that time.Unix takes after roughly 922
	// years past 1601, which a 64-bit FILETIME can express more than sixty times
	// over, so the arithmetic has to stay in seconds to cover the whole range.
	secondsSince1601 := ticks / ticksPerSecond
	remainderTicks := ticks % ticksPerSecond

	secondsFromUnixEpoch := int64(secondsSince1601) - int64(ticksBetween1601AndUnix/ticksPerSecond)

	// The result is normalised to UTC to match SetTime, which stores a UTC value:
	// time.Unix returns a Time in the machine's local zone, which would otherwise
	// make the zone of a DateTime depend on how it was populated, and Describe
	// prints these values under a label that says UTC.
	dt.time = time.Unix(secondsFromUnixEpoch, int64(remainderTicks)*100).UTC()
}

// GetTicks returns the number of 100-nanosecond intervals (ticks) stored in the DateTime instance.
//
// Returns:
// - A uint64 value representing the number of 100-nanosecond intervals since 1601-01-01 00:00:00 UTC.
func (dt *DateTime) GetTicks() uint64 {
	return dt.ticks
}

// SetTime sets the time for the DateTime instance.
//
// Parameters:
// - t: A time.Time value to set.
//
// Note:
// This function updates both the Time field and recalculates the corresponding Ticks field based on the provided time.
func (dt *DateTime) SetTime(t time.Time) {
	epoch1601 := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	if t.UTC().Equal(epoch1601) {
		dt.time = epoch1601
		dt.ticks = 0
		return
	}

	truncated := t.UTC().Truncate(100 * time.Nanosecond)
	dt.time = truncated
	nanos := truncated.UnixNano()
	if nanos >= 0 {
		dt.ticks = ticksBetween1601AndUnix + uint64(nanos/100)
	} else {
		// Truncated ensures nanos is divisible by 100
		dt.ticks = ticksBetween1601AndUnix - uint64((-nanos)/100)
	}
}

// GetTime returns the time.Time value stored in the DateTime instance.
//
// Returns:
// - A time.Time value representing the stored time.
func (dt *DateTime) GetTime() time.Time {
	return dt.time
}

// ToTicks returns the number of 100-nanosecond intervals (ticks) that have elapsed since 1601-01-01 00:00:00 UTC.
//
// Returns:
// - A uint64 value representing the number of 100-nanosecond intervals (ticks) since 1601-01-01 00:00:00 UTC.
//
// Note:
// This function provides a way to retrieve the internal tick count of the DateTime instance, which is useful for binary time representations and calculations.
func (dt DateTime) ToTicks() uint64 {
	return dt.ticks
}

// String returns the string representation of the DateTime instance.
//
// Returns:
// - A string representing the DateTime instance in the default format used by the time.Time type.
//
// Note:
// This function leverages the String method of the embedded time.Time type to provide a human-readable
// representation of the DateTime instance. The format typically includes the date, time, and time zone.
func (dt DateTime) String() string {
	return dt.time.String()
}

// Marshal converts the DateTime instance to its binary representation in little-endian format.
//
// Returns:
// - A byte slice containing the binary representation of the DateTime instance in little-endian format, otherwise an error.
//
// Note:
// This function encodes the Ticks field of the DateTime instance as a 64-bit unsigned integer in little-endian format.
// The function returns a byte slice containing the binary representation of the DateTime instance in little-endian format.
func (dt DateTime) Marshal() ([]byte, error) {
	return binary.LittleEndian.AppendUint64(nil, dt.ticks), nil
}

// Unmarshal converts the binary representation of the DateTime instance to a DateTime object.
//
// Parameters:
// - data: A byte slice containing the binary representation of the DateTime instance.
//
// Returns:
// - An error if the unmarshalling fails, otherwise nil.
//
// Note:
// This function decodes the Ticks field of the DateTime instance from a 64-bit unsigned integer in little-endian format.
// The function expects the data to be a 64-bit unsigned integer in little-endian format.
func (dt *DateTime) Unmarshal(data []byte) error {
	if len(data) != 8 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	// Routed through SetTicks so the derived time field is recomputed with the tick
	// count: assigning dt.ticks alone leaves the two fields describing different
	// instants, and every other mutator in this file maintains the pair.
	dt.SetTicks(binary.LittleEndian.Uint64(data))

	return nil
}
