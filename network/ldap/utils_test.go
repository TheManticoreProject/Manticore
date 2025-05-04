package ldap_test

import (
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/ldap"
)

func TestConvertSecondsToLDAPDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"One day", 86400, "864000000000"},
		{"One hour", 3600, "36000000000"},
		{"One minute", 60, "600000000"},
		{"One second", 1, "10000000"},
		{"Zero", 0, "0"},
		{"Negative value", -1, "-10000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ldap.ConvertSecondsToLDAPDuration(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertSecondsToLDAPDuration(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertUnixTimeStampToLDAPTimeStamp(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected int64
	}{
		{"Epoch", time.Unix(0, 0), ldap.UnixTimestampStart},
		{"One second after epoch", time.Unix(1, 0), ldap.UnixTimestampStart + 10000000},
		{"One day after epoch", time.Unix(86400, 0), 116445600000000000},
		{"Current time", time.Now(), ldap.ConvertUnixTimeStampToLDAPTimeStamp(time.Now())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ldap.ConvertUnixTimeStampToLDAPTimeStamp(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertUnixTimeStampToLDAPTimeStamp(%v) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}
