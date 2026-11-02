package utils_test

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/source"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/utils"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

// TestSetTimeBefore1601DoesNotWrap checks a time before the FILETIME epoch gets a
// tick count of 0 rather than wrapping to a date in the far future, keeps its own
// time, and that a time past the 64-bit FILETIME range is clamped to its maximum.
func TestSetTimeBefore1601DoesNotWrap(t *testing.T) {
	for _, tc := range []struct {
		in    time.Time
		ticks uint64
	}{
		{time.Time{}, 0}, // 0001-01-01, the zero time.Time
		{time.Date(1500, 6, 1, 12, 0, 0, 0, time.UTC), 0},
		{time.Date(1600, 12, 31, 23, 59, 59, 999999900, time.UTC), 0},
		{time.Date(1601, 1, 1, 0, 0, 0, 100, time.UTC), 1},
		{time.Date(70000, 1, 1, 0, 0, 0, 0, time.UTC), math.MaxUint64}, // past FILETIME's range (year 60056)
	} {
		var dt utils.DateTime
		dt.SetTime(tc.in)
		if got := dt.GetTicks(); got != tc.ticks {
			t.Errorf("SetTime(%s): ticks = %d, want %d", tc.in, got, tc.ticks)
		}
		if !dt.GetTime().Equal(tc.in) {
			t.Errorf("SetTime(%s): time = %s, want it kept", tc.in, dt.GetTime())
		}
	}
}

// TestZeroTimeEncodesAsZeroFileTime checks a zero time.Time written as an AD
// KeyCredentialLink timestamp is the zero FILETIME, not 000089dde831fef8.
func TestZeroTimeEncodesAsZeroFileTime(t *testing.T) {
	got := utils.ConvertToBinaryTime(time.Time{}, source.KeySource{Value: source.KeySource_AD},
		version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2})
	if !bytes.Equal(got, make([]byte, 8)) {
		t.Errorf("ConvertToBinaryTime(zero) = %x, want 0000000000000000", got)
	}
}
