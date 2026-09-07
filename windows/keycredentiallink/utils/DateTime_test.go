package utils_test

import (
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/utils"
)

// ticksFromTime computes 100ns ticks since 1601-01-01 UTC without relying on
// time.Time values outside the int64 nanosecond range (avoids overflow).
func ticksFromTime(t time.Time) uint64 {
	const ticksBetween1601AndUnix uint64 = 116444736000000000
	trunc := t.UTC().Truncate(100 * time.Nanosecond)
	nanos := trunc.UnixNano()
	if nanos >= 0 {
		return ticksBetween1601AndUnix + uint64(nanos/100)
	}
	return ticksBetween1601AndUnix - uint64((-nanos)/100)
}

func TestNewDateTime(t *testing.T) {
	t.Run("with zero ticks should set current time", func(t *testing.T) {
		before := time.Now()
		dt := utils.NewDateTimeFromTicks(0)
		after := time.Now()

		if !(dt.GetTime().After(before) || dt.GetTime().Equal(before)) {
			t.Error("Expected time to be after or equal to before time")
		}
		if !(dt.GetTime().Before(after) || dt.GetTime().Equal(after)) {
			t.Error("Expected time to be before or equal to after time")
		}
		if dt.GetTicks() == 0 {
			t.Error("Expected non-zero ticks")
		}
	})

	t.Run("with specific ticks should set correct time", func(t *testing.T) {
		// 132918192030000000 ticks = 2022-03-15 12:00:03 UTC
		dt := utils.NewDateTimeFromTicks(uint64(132918192030000000))

		expected := time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC)
		if !dt.GetTime().UTC().Equal(expected) {
			t.Errorf("Expected time %v, got %v", expected, dt.GetTime().UTC())
		}
		if dt.GetTicks() != uint64(132918192030000000) {
			t.Errorf("Expected ticks %d, got %d", uint64(132918192030000000), dt.GetTicks())
		}
	})
}

func TestDateTime_ToUniversalTime(t *testing.T) {
	dt := utils.NewDateTimeFromTicks(uint64(132918192030000000)) // 2022-03-15 12:00:03 UTC
	utc := dt.ToUniversalTime()

	expected := time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC)
	if !utc.Equal(expected) {
		t.Errorf("Expected UTC time %v, got %v", expected, utc)
	}
}

func TestDateTime_ToTicks(t *testing.T) {
	ticks := uint64(132918192030000000)
	dt := utils.NewDateTimeFromTicks(ticks)

	if dt.ToTicks() != ticks {
		t.Errorf("Expected ticks %d, got %d", ticks, dt.ToTicks())
	}
}

func TestDateTime_String(t *testing.T) {
	dt := utils.NewDateTimeFromTicks(uint64(132918192030000000)) // 2022-03-15 12:00:03 UTC
	expected := time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC)

	// Compare the actual time values instead of string representations to avoid timezone issues
	if !dt.GetTime().UTC().Equal(expected) {
		t.Errorf("Expected time %v, got %v", expected, dt.GetTime().UTC())
	}
}

func TestDateTime_Marshal(t *testing.T) {
	dt := utils.NewDateTimeFromTicks(uint64(132918192030000000))
	data, err := dt.Marshal()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(data) != 8 { // Should be 8 bytes for uint64
		t.Errorf("Expected data length 8, got %d", len(data))
	}
}

func TestDateTime_Unmarshal(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		dt := &utils.DateTime{}
		// 132918192030000000 in little-endian = 2022-03-15 12:00:03 UTC
		data := []byte{0x80, 0xa3, 0x22, 0x34, 0x64, 0x38, 0xd8, 0x01}
		err := dt.Unmarshal(data)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if dt.GetTicks() != uint64(132918192030000000) {
			t.Errorf("Expected ticks %d, got %d", uint64(132918192030000000), dt.GetTicks())
		}
	})

	t.Run("invalid data length", func(t *testing.T) {
		dt := &utils.DateTime{}
		data := []byte{0x30, 0x75, 0x20} // Too short

		err := dt.Unmarshal(data)

		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "invalid data length: 3" {
			t.Errorf("Expected error message 'invalid data length: 3', got '%s'", err.Error())
		}
	})
}

func TestDateTime_GetTime_SetTime(t *testing.T) {
	currentTime := time.Now().UTC().Truncate(100 * time.Nanosecond)

	testCases := []struct {
		name     string
		setTime  time.Time
		expected time.Time
	}{
		{
			name:     "Current time",
			setTime:  currentTime,
			expected: currentTime,
		},
		{
			name:     "Past time",
			setTime:  time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Future time",
			setTime:  time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		{
			name:     "Epoch time",
			setTime:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dt := utils.NewDateTimeFromTicks(0)
			dt.SetTime(tc.setTime)
			result := dt.GetTime()

			if !result.Equal(tc.expected) {
				t.Errorf("Expected time %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestDateTime_GetTicks_SetTicks(t *testing.T) {
	testCases := []struct {
		name        string
		setTicks    uint64
		expectedErr bool
	}{
		{
			name:        "Typical value",
			setTicks:    uint64(132918192030000000), // 2022-03-15 12:00:03 UTC
			expectedErr: false,
		},
		{
			name:        "Maximum value",
			setTicks:    ^uint64(0), // Max uint64
			expectedErr: false,
		},
		{
			name:        "Early date",
			setTicks:    uint64(116444736000000000), // 1970-01-01 00:00:00 UTC
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dt := utils.NewDateTimeFromTicks(0)
			dt.SetTicks(tc.setTicks)
			result := dt.GetTicks()

			if result != tc.setTicks {
				t.Errorf("Expected ticks %d, got %d", tc.setTicks, result)
			}

			// Verify that the time representation is consistent
			dt2 := utils.NewDateTimeFromTicks(tc.setTicks)
			if dt2.GetTicks() != result {
				t.Errorf("Tick values not consistent after recreation: expected %d, got %d",
					result, dt2.GetTicks())
			}
		})
	}
}

func TestNewDateTimeFromTime(t *testing.T) {
	epoch1601 := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	unixEpoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	// nanosecondsBetween1601AndEpoch is used for calculating expected ticks from time.Time
	// and is consistent with the definition of ticks in DateTime.go
	// (number of 100-nanosecond intervals since 1601-01-01 00:00:00 UTC)
	// Note: The current implementation of NewDateTimeFromTime might not correctly use this epoch for tick calculation,
	// leading to test failures which would indicate a bug.
	nanosecondsBetween1601AndEpoch := uint64(unixEpoch.UnixNano() - epoch1601.UnixNano())

	testCases := []struct {
		name          string
		inputTime     time.Time
		expectedTime  time.Time // Expected time after truncation to 100ns
		expectedTicks uint64    // Expected ticks based on 1601-01-01 UTC epoch
	}{
		{
			name:          "Current time with nanosecond precision",
			inputTime:     time.Date(2023, 10, 27, 10, 30, 45, 123456789, time.UTC),
			expectedTime:  time.Date(2023, 10, 27, 10, 30, 45, 123456700, time.UTC), // Truncated to 100ns
			expectedTicks: ticksFromTime(time.Date(2023, 10, 27, 10, 30, 45, 123456700, time.UTC)),
		},
		{
			name:          "Time with exact 100-nanosecond precision",
			inputTime:     time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC), // Corresponds to 132918192030000000 ticks
			expectedTime:  time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC),
			expectedTicks: uint64(132918192030000000),
		},
		{
			name:          "Unix epoch time (1970-01-01 00:00:00 UTC)",
			inputTime:     unixEpoch,
			expectedTime:  unixEpoch,
			expectedTicks: nanosecondsBetween1601AndEpoch / 100, // Ticks from 1601 to 1970
		},
		{
			name:          "Time before Unix epoch (e.g., 1800-01-01 00:00:00 UTC)",
			inputTime:     time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedTime:  time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedTicks: ticksFromTime(time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		{
			name:          "1601 epoch time (1601-01-01 00:00:00 UTC)",
			inputTime:     epoch1601,
			expectedTime:  epoch1601,
			expectedTicks: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dt := utils.NewDateTimeFromTime(tc.inputTime)

			// Verify that the time component is correctly set and truncated
			if !dt.GetTime().UTC().Equal(tc.expectedTime) {
				t.Errorf("Expected time %v, got %v", tc.expectedTime, dt.GetTime().UTC())
			}

			// Verify that the ticks component is correctly calculated based on 1601 epoch
			if dt.GetTicks() != tc.expectedTicks {
				t.Errorf("Expected ticks %d, got %d", tc.expectedTicks, dt.GetTicks())
			}
		})
	}
}

// A 64-bit FILETIME can express dates far beyond what a single scaling to
// nanoseconds can hold, so large tick counts have to decode to their real date
// instead of a wrapped one. Before the fix the year-9999 value below decoded to
// 1816, and the largest signed and unsigned values — 9.2 billion seconds apart —
// decoded to the same instant.
func TestDateTime_SetTicks_LargeValues(t *testing.T) {
	testCases := []struct {
		name         string
		ticks        uint64
		expectedYear int
	}{
		{name: "1601 epoch", ticks: 0, expectedYear: 1601},
		{name: "typical value", ticks: 132918192030000000, expectedYear: 2022},
		{name: "9999-12-31 (the largest date .NET represents)", ticks: 2650467743999999999, expectedYear: 9999},
		{name: "maximum signed value", ticks: 0x7FFFFFFFFFFFFFFF, expectedYear: 30828},
		{name: "maximum unsigned value", ticks: 0xFFFFFFFFFFFFFFFF, expectedYear: 60056},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dt := utils.DateTime{}
			dt.SetTicks(tc.ticks)

			if got := dt.GetTime().UTC().Year(); got != tc.expectedYear {
				t.Errorf("SetTicks(%d) year = %d, want %d (time %s)", tc.ticks, got, tc.expectedYear, dt.GetTime().UTC())
			}

			if dt.GetTicks() != tc.ticks {
				t.Errorf("GetTicks() = %d, want %d", dt.GetTicks(), tc.ticks)
			}
		})
	}
}

// Distinct tick counts must decode to distinct instants; a wrapped multiplication
// mapped the largest signed and unsigned values onto the same one.
func TestDateTime_SetTicks_DistinctValuesStayDistinct(t *testing.T) {
	first := utils.DateTime{}
	first.SetTicks(0x7FFFFFFFFFFFFFFF)

	second := utils.DateTime{}
	second.SetTicks(0xFFFFFFFFFFFFFFFF)

	if first.GetTime().Equal(second.GetTime()) {
		t.Errorf("two tick counts 9.2e18 apart decoded to the same instant: %s", first.GetTime().UTC())
	}
}

// The sub-second remainder is carried separately from the whole seconds, so it has
// to survive the conversion.
func TestDateTime_SetTicks_SubSecondPrecision(t *testing.T) {
	// 2022-03-15 12:00:03 UTC plus 1234567 ticks (0.1234567s)
	dt := utils.DateTime{}
	dt.SetTicks(132918192030000000 + 1234567)

	expected := time.Date(2022, 3, 15, 12, 0, 3, 123456700, time.UTC)
	if !dt.GetTime().UTC().Equal(expected) {
		t.Errorf("GetTime() = %s, want %s", dt.GetTime().UTC(), expected)
	}
}

// withNonUTCLocalZone points time.Local at a fixed offset for the duration of a
// test. The defect these tests cover is invisible on a UTC host — which is what CI
// runs — because the local zone and UTC then coincide, so the zone has to be pinned
// for the assertions to mean anything.
func withNonUTCLocalZone(t *testing.T) {
	t.Helper()

	original := time.Local
	time.Local = time.FixedZone("TEST", 2*60*60)
	t.Cleanup(func() { time.Local = original })
}

// A DateTime must report the same zone however it was populated: SetTime stores a
// UTC value, and time.Unix inside SetTicks returns a local-zone one, so a decoded
// timestamp used to print with the collector's offset under Describe's "(UTC)" label.
func TestDateTime_SetTicks_StoresUTC(t *testing.T) {
	withNonUTCLocalZone(t)

	dt := utils.DateTime{}
	dt.SetTicks(132918192030000000) // 2022-03-15 12:00:03 UTC

	if zone, offset := dt.GetTime().Zone(); offset != 0 {
		t.Errorf("SetTicks stored zone %s with offset %d, want UTC (offset 0)", zone, offset)
	}

	if got, want := dt.String(), "2022-03-15 12:00:03 +0000 UTC"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// SetTicks and SetTime must agree on the zone for the same instant.
func TestDateTime_SetTicksAndSetTimeAgreeOnZone(t *testing.T) {
	withNonUTCLocalZone(t)

	viaTicks := utils.DateTime{}
	viaTicks.SetTicks(132918192030000000)

	viaTime := utils.NewDateTimeFromTime(viaTicks.GetTime())

	if viaTicks.String() != viaTime.String() {
		t.Errorf("SetTicks gave %q, SetTime gave %q for the same instant", viaTicks.String(), viaTime.String())
	}
}

// Unmarshal has to leave the DateTime in the same state as SetTicks with the same
// tick count: assigning only the ticks left the time accessors reporting the zero
// time for a perfectly valid timestamp.
func TestDateTime_Unmarshal_SetsTheTime(t *testing.T) {
	data := []byte{0x80, 0xa3, 0x22, 0x34, 0x64, 0x38, 0xd8, 0x01} // 2022-03-15 12:00:03 UTC

	dt := utils.DateTime{}
	if err := dt.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}

	expected := time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC)
	if !dt.GetTime().UTC().Equal(expected) {
		t.Errorf("GetTime() = %s, want %s", dt.GetTime().UTC(), expected)
	}

	if dt.GetTicks() != 132918192030000000 {
		t.Errorf("GetTicks() = %d, want 132918192030000000", dt.GetTicks())
	}

	// The two fields must describe the same instant, whichever mutator was used.
	viaSetTicks := utils.DateTime{}
	viaSetTicks.SetTicks(dt.GetTicks())
	if !dt.GetTime().Equal(viaSetTicks.GetTime()) {
		t.Errorf("Unmarshal gave time %s, SetTicks gave %s for the same ticks", dt.GetTime(), viaSetTicks.GetTime())
	}
}
