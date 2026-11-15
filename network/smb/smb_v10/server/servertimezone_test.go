package server

import (
	"testing"
	"time"
)

// TestServerTimeZoneBias checks the SMB1 NEGOTIATE ServerTimeZone is the host's
// real UTC bias in minutes, not a constant 0. A hard-coded 0 made every server
// look like a UTC host, which fingerprints it against hosts whose other evidence
// shows a non-UTC zone. The value is the minutes to add to local time to reach
// UTC ([MS-CIFS] 2.2.4.52.2).
func TestServerTimeZoneBias(t *testing.T) {
	cases := []struct {
		name          string
		offsetSeconds int
		want          int16
	}{
		{"UTC", 0, 0},
		{"UTC+1 (CET)", 1 * 3600, -60},
		{"UTC+2 (CEST)", 2 * 3600, -120},
		{"UTC-5 (EST)", -5 * 3600, 300},
		{"UTC+5:30 (IST)", 5*3600 + 30*60, -330},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			zone := time.FixedZone(tc.name, tc.offsetSeconds)
			got := serverTimeZoneBias(time.Date(2026, 6, 1, 12, 0, 0, 0, zone))
			if int16(got) != tc.want {
				t.Errorf("serverTimeZoneBias(%s) = %d, want %d", tc.name, int16(got), tc.want)
			}
		})
	}
}
