package ldap_attributes_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

// msPKI-Enrollment-Flag is a bit field and a certificate template almost always sets
// several flags at once. An exact-match lookup reported those ordinary values as
// UnknownEnrollmentFlag(<decimal>) even though every flag in them is named.
func TestMSPKIEnrollmentFlag_String(t *testing.T) {
	testCases := []struct {
		name     string
		value    ldap_attributes.MSPKIEnrollmentFlag
		expected string
	}{
		{name: "no flags", value: 0x00000000, expected: ""},
		{name: "single flag", value: 0x00000020, expected: "Auto Enrollment"},
		{
			name:     "auto enrollment and publish to DS",
			value:    0x00000028,
			expected: "Auto Enrollment|Publish to DS",
		},
		{
			name:     "no security extension alone",
			value:    0x00080000,
			expected: "No Security Extension",
		},
		{
			// PUBLISH_TO_DS | AUTO_ENROLLMENT | USER_INTERACTION_REQUIRED |
			// ISSUANCE_POLICIES_FROM_REQUEST. Note 0x80 and 0x200 are not assigned by
			// MS-CRTD, so a realistic value skips them.
			name:  "a realistic template flag set",
			value: 0x00020128,
			expected: "Auto Enrollment|Issuance Policies From Request|Publish to DS|" +
				"User Interaction Required",
		},
		{
			name:     "unnamed bit alone",
			value:    0x80000000,
			expected: "UnknownEnrollmentFlag(0x80000000)",
		},
		{
			name:     "named flag combined with an unnamed bit",
			value:    0x80000020,
			expected: "Auto Enrollment|UnknownEnrollmentFlag(0x80000000)",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.value.String(); got != testCase.expected {
				t.Errorf("MSPKIEnrollmentFlag(0x%08x).String() = %q, want %q", uint32(testCase.value), got, testCase.expected)
			}
		})
	}
}

// An unrecognised bit must not hide the flags that are recognised.
func TestMSPKIEnrollmentFlag_UnknownBitDoesNotHideKnownFlags(t *testing.T) {
	got := ldap_attributes.MSPKIEnrollmentFlag(0x00000028 | 0x40000000).String()

	if !strings.Contains(got, "Auto Enrollment") || !strings.Contains(got, "Publish to DS") {
		t.Errorf("String() = %q, want it to still name the recognised flags", got)
	}

	if !strings.Contains(got, "UnknownEnrollmentFlag(0x40000000)") {
		t.Errorf("String() = %q, want it to report the unrecognised bit", got)
	}
}

// GetFlags decomposes the value into the individual named flags it carries.
func TestMSPKIEnrollmentFlag_GetFlags(t *testing.T) {
	flags := ldap_attributes.MSPKIEnrollmentFlag(0x00000028).GetFlags()

	expected := []ldap_attributes.MSPKIEnrollmentFlag{
		ldap_attributes.MSPKI_ENROLLMENT_FLAG_PUBLISH_TO_DS,
		ldap_attributes.MSPKI_ENROLLMENT_FLAG_AUTO_ENROLLMENT,
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
func TestMSPKIEnrollmentFlag_EveryFlagIsNamed(t *testing.T) {
	for flag, name := range ldap_attributes.MSPKIEnrollmentFlagMap {
		if got := flag.String(); got != name {
			t.Errorf("MSPKIEnrollmentFlag(0x%08x).String() = %q, want %q", uint32(flag), got, name)
		}
	}
}
