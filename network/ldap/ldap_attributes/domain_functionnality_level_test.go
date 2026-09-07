package ldap_attributes_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

// IsSupported must agree with the table of defined levels. It used to restate that
// table as a chain of equality comparisons and omitted 2003 Interim, so level 1 was
// the only defined level reporting false.
func TestDomainFunctionalityLevel_IsSupportedMatchesTheTable(t *testing.T) {
	for level, name := range ldap_attributes.DomainFunctionalityLevelToWindowsVersion {
		if !level.IsSupported() {
			t.Errorf("level %d (%s) is named by the table but IsSupported() = false", level, name)
		}
	}
}

// Every defined level, checked explicitly so a level dropped from the table is caught
// rather than silently skipped by the loop above.
func TestDomainFunctionalityLevel_EveryDefinedLevel(t *testing.T) {
	testCases := []struct {
		name  string
		level ldap_attributes.DomainFunctionalityLevel
	}{
		{name: "2000", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2000},
		{name: "2003 Interim", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2003_INTERIM},
		{name: "2003", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2003},
		{name: "2008", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2008},
		{name: "2008 R2", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2008_R2},
		{name: "2012", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2012},
		{name: "2012 R2", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2012_R2},
		{name: "2016", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2016},
		{name: "2025", level: ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2025},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if !testCase.level.IsSupported() {
				t.Errorf("DOMAIN_FUNCTIONALITY_LEVEL_%s (%d).IsSupported() = false, want true", testCase.name, testCase.level)
			}

			if got := testCase.level.String(); got == "" {
				t.Errorf("level %d has an empty String()", testCase.level)
			}
		})
	}
}

// Values the package does not define must report false and render as unknown.
func TestDomainFunctionalityLevel_UndefinedLevels(t *testing.T) {
	for _, level := range []ldap_attributes.DomainFunctionalityLevel{8, 9, 11, 255} {
		t.Run(ldap_attributes.DomainFunctionalityLevel(level).String(), func(t *testing.T) {
			if level.IsSupported() {
				t.Errorf("level %d.IsSupported() = true, want false", level)
			}

			if got, want := level.String(), "UNKNOWN"; len(got) < len(want) || got[:len(want)] != want {
				t.Errorf("level %d.String() = %q, want it to start with %q", level, got, want)
			}
		})
	}
}

// The 2003 Interim level, which is what the omission affected.
func TestDomainFunctionalityLevel_2003InterimIsSupported(t *testing.T) {
	level := ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2003_INTERIM

	if level != 1 {
		t.Fatalf("DOMAIN_FUNCTIONALITY_LEVEL_2003_INTERIM = %d, want 1", level)
	}

	if !level.IsSupported() {
		t.Errorf("IsSupported() = false for 2003 Interim, which String() names %q", level.String())
	}
}

// Ordering is used by the comparison helpers on Domain, so it has to be monotonic.
func TestDomainFunctionalityLevel_Ordering(t *testing.T) {
	if !ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2003.Less(ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2016) {
		t.Error("2003 should be less than 2016")
	}

	if ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2025.Less(ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2016) {
		t.Error("2025 should not be less than 2016")
	}

	if !ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2016.Equal(ldap_attributes.DOMAIN_FUNCTIONALITY_LEVEL_2016) {
		t.Error("2016 should equal itself")
	}
}
