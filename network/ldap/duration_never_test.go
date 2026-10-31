package ldap_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap"
)

// TestConvertLDAPDurationToSecondsNever checks the Active Directory "never" value
// (math.MinInt64, as stored in maxPwdAge or lockoutDuration) converts to a positive
// number of seconds instead of overflowing on negation, alongside ordinary negative
// and positive durations.
func TestConvertLDAPDurationToSecondsNever(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{strconv.FormatInt(math.MinInt64, 10), 922337203685},
		{strconv.FormatInt(math.MaxInt64, 10), 922337203685},
		{"-36288000000000", 3628800},
		{"864000000000", 86400},
		{"-1", 0},
		{"", 0},
	}

	for _, tt := range tests {
		result, err := ldap.ConvertLDAPDurationToSeconds(tt.input)
		if err != nil {
			t.Fatalf("ConvertLDAPDurationToSeconds(%q) error = %v", tt.input, err)
		}
		if result != tt.expected {
			t.Errorf("ConvertLDAPDurationToSeconds(%q) = %d; want %d", tt.input, result, tt.expected)
		}
	}
}
