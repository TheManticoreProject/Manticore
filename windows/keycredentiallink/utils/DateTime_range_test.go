package utils_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/source"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/utils"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

// TestDateTime_SetTime_OutsideUnixNanoRange checks that dates outside the roughly
// 1678-2262 span representable by UnixNano() encode to the correct tick count and
// round-trip back to the same instant.
func TestDateTime_SetTime_OutsideUnixNanoRange(t *testing.T) {
	testCases := []struct {
		name  string
		time  time.Time
		ticks uint64
	}{
		// 1650-01-01 is 49 years (12 leap days) after 1601-01-01.
		{"1650", time.Date(1650, 1, 1, 0, 0, 0, 0, time.UTC), uint64(49*365+12) * 86400 * 10000000},
		// 3000-01-01 is 510974 days after 1601-01-01.
		{"3000", time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC), 441481536000000000},
		// 9999-12-31 23:59:59.9999999 is the .NET DateTime.MaxValue as a FILETIME.
		{"9999", time.Date(9999, 12, 31, 23, 59, 59, 999999900, time.UTC), 2650467743999999999},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dt := utils.NewDateTimeFromTime(tc.time)
			if got := dt.GetTicks(); got != tc.ticks {
				t.Fatalf("ticks: got %d want %d", got, tc.ticks)
			}

			raw, err := dt.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var back utils.DateTime
			if err := back.Unmarshal(raw); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !back.GetTime().Equal(tc.time) {
				t.Fatalf("round-trip: got %s want %s", back.GetTime(), tc.time)
			}

			ksrc := source.KeySource{Value: source.KeySource_AD}
			kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}
			encoded := utils.ConvertToBinaryTime(tc.time, ksrc, kcv)
			if got := binary.LittleEndian.Uint64(encoded); got != tc.ticks {
				t.Fatalf("ConvertToBinaryTime: got %d want %d", got, tc.ticks)
			}
			decoded, err := utils.ConvertFromBinaryTime(encoded, ksrc, kcv)
			if err != nil {
				t.Fatalf("ConvertFromBinaryTime: %v", err)
			}
			if !decoded.GetTime().Equal(tc.time) {
				t.Fatalf("ConvertFromBinaryTime: got %s want %s", decoded.GetTime(), tc.time)
			}
		})
	}
}
