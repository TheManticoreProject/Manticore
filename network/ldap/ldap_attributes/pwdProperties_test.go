package ldap_attributes_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

// pwdProperties is a bit field, so every flag set in the value has to be named. An
// exact-match lookup returned an empty string for any value with more than one flag,
// which is the normal shape of a domain policy.
func TestPasswordProperties_String(t *testing.T) {
	testCases := []struct {
		name     string
		value    ldap_attributes.PasswordProperties
		expected string
	}{
		{name: "no flags", value: 0x00000000, expected: ""},
		{name: "single flag", value: 0x00000001, expected: "DOMAIN_PASSWORD_COMPLEX"},
		{
			name:     "complex and store cleartext",
			value:    0x00000011,
			expected: "DOMAIN_PASSWORD_COMPLEX|DOMAIN_PASSWORD_STORE_CLEARTEXT",
		},
		{
			name:     "complex, lockout admins and store cleartext",
			value:    0x00000019,
			expected: "DOMAIN_LOCKOUT_ADMINS|DOMAIN_PASSWORD_COMPLEX|DOMAIN_PASSWORD_STORE_CLEARTEXT",
		},
		{
			name:  "every defined flag",
			value: 0x0000003F,
			expected: "DOMAIN_LOCKOUT_ADMINS|DOMAIN_PASSWORD_COMPLEX|DOMAIN_PASSWORD_NO_ANON_CHANGE|" +
				"DOMAIN_PASSWORD_NO_CLEAR_CHANGE|DOMAIN_PASSWORD_STORE_CLEARTEXT|DOMAIN_REFUSE_PASSWORD_CHANGE",
		},
		{name: "only undefined bits", value: 0x00000040, expected: ""},
		{name: "known flag with an undefined bit", value: 0x00000041, expected: "DOMAIN_PASSWORD_COMPLEX"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.value.String(); got != testCase.expected {
				t.Errorf("PasswordProperties(0x%08x).String() = %q, want %q", uint32(testCase.value), got, testCase.expected)
			}
		})
	}
}

// Description has the same bit-field shape as String and must describe every set flag.
func TestPasswordProperties_Description(t *testing.T) {
	t.Run("single flag", func(t *testing.T) {
		got := ldap_attributes.PasswordProperties(0x00000001).Description()
		if !strings.Contains(got, "mix of at least two") {
			t.Errorf("Description() = %q, want the complexity description", got)
		}
		if strings.Contains(got, "\n") {
			t.Errorf("Description() = %q, want a single line for a single flag", got)
		}
	})

	t.Run("two flags", func(t *testing.T) {
		got := ldap_attributes.PasswordProperties(0x00000011).Description()

		lines := strings.Split(got, "\n")
		if len(lines) != 2 {
			t.Fatalf("Description() returned %d lines, want 2:\n%s", len(lines), got)
		}
		if !strings.Contains(got, "mix of at least two") {
			t.Errorf("Description() is missing the complexity description:\n%s", got)
		}
		if !strings.Contains(got, "plaintext password for all users") {
			t.Errorf("Description() is missing the cleartext-storage description:\n%s", got)
		}
	})

	t.Run("no flags", func(t *testing.T) {
		if got := ldap_attributes.PasswordProperties(0).Description(); got != "" {
			t.Errorf("Description() = %q, want an empty string", got)
		}
	})
}

// GetFlags decomposes the value into the individual flags it carries.
func TestPasswordProperties_GetFlags(t *testing.T) {
	flags := ldap_attributes.PasswordProperties(0x00000011).GetFlags()

	expected := []ldap_attributes.PasswordProperties{
		ldap_attributes.PASSWORD_PROPERTY_DOMAIN_PASSWORD_COMPLEX,
		ldap_attributes.PASSWORD_PROPERTY_DOMAIN_PASSWORD_STORE_CLEARTEXT,
	}

	if len(flags) != len(expected) {
		t.Fatalf("GetFlags() returned %d flags (%v), want %d", len(flags), flags, len(expected))
	}

	for i := range expected {
		if flags[i] != expected[i] {
			t.Errorf("GetFlags()[%d] = 0x%08x, want 0x%08x", i, uint32(flags[i]), uint32(expected[i]))
		}
	}
}

// Every name in the map must be reachable through String() on its own flag.
func TestPasswordProperties_EveryFlagIsNamed(t *testing.T) {
	for flag, name := range ldap_attributes.PasswordPropertiesMap {
		if got := flag.String(); got != name {
			t.Errorf("PasswordProperties(0x%08x).String() = %q, want %q", uint32(flag), got, name)
		}

		if flag.Description() == "" {
			t.Errorf("PasswordProperties(0x%08x) (%s) has no description", uint32(flag), name)
		}
	}
}
