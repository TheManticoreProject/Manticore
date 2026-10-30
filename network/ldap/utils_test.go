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

func TestConvertLDAPTimeStampToUnixTimeStamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"Empty value", "", 0},
		{"Epoch", "116444736000000000", 0},
		{"One day after epoch", "116445600000000000", 86400},
		{"Known time", "133444736000000000", 1700000000},
		{"Before epoch (year 1601)", "0", 0},
		// 0x7FFFFFFFFFFFFFFF "never" sentinel: must not overflow int64.
		{"Max int64 never sentinel", "9223372036854775807", 910692730085},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ldap.ConvertLDAPTimeStampToUnixTimeStamp(tt.input)
			if err != nil {
				t.Fatalf("ConvertLDAPTimeStampToUnixTimeStamp(%q) returned error: %s", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ConvertLDAPTimeStampToUnixTimeStamp(%q) = %d; want %d", tt.input, result, tt.expected)
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

func TestGetDomainFromDistinguishedName(t *testing.T) {
	tests := []struct {
		dn   string
		want string
	}{
		{"CN=John Doe,OU=Users,DC=example,DC=com", "example.com"},
		{"dc=corp,dc=local", "corp.local"},
		{"CN=x, DC=corp, DC=local", "corp.local"},
		{"CN=Smith\\, DC=fake,OU=Users,DC=corp,DC=local", "corp.local"},
		{`CN=foo\0ADEL:0f5c1a7d-3b4e-4f6a-9c2d-1e2f3a4b5c6d,CN=Deleted Objects,DC=corp,DC=local`, "corp.local"},
		{"CN=Users", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := ldap.GetDomainFromDistinguishedName(tt.dn); got != tt.want {
			t.Errorf("GetDomainFromDistinguishedName(%q) = %q, want %q", tt.dn, got, tt.want)
		}
	}
}
