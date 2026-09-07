package ldap_attributes

// Domain Functionality Level
// Src: https://eightwone.com/references/ad-functional-levels/

import "fmt"

type DomainFunctionalityLevel uint8

const (
	DOMAIN_FUNCTIONALITY_LEVEL_2000         DomainFunctionalityLevel = 0
	DOMAIN_FUNCTIONALITY_LEVEL_2003_INTERIM DomainFunctionalityLevel = 1
	DOMAIN_FUNCTIONALITY_LEVEL_2003         DomainFunctionalityLevel = 2
	DOMAIN_FUNCTIONALITY_LEVEL_2008         DomainFunctionalityLevel = 3
	DOMAIN_FUNCTIONALITY_LEVEL_2008_R2      DomainFunctionalityLevel = 4
	DOMAIN_FUNCTIONALITY_LEVEL_2012         DomainFunctionalityLevel = 5
	DOMAIN_FUNCTIONALITY_LEVEL_2012_R2      DomainFunctionalityLevel = 6
	DOMAIN_FUNCTIONALITY_LEVEL_2016         DomainFunctionalityLevel = 7
	DOMAIN_FUNCTIONALITY_LEVEL_2025         DomainFunctionalityLevel = 10
)

var DomainFunctionalityLevelToWindowsVersion = map[DomainFunctionalityLevel]string{
	DOMAIN_FUNCTIONALITY_LEVEL_2000:         "Windows 2000",
	DOMAIN_FUNCTIONALITY_LEVEL_2003_INTERIM: "Windows Server 2003 Interim",
	DOMAIN_FUNCTIONALITY_LEVEL_2003:         "Windows Server 2003",
	DOMAIN_FUNCTIONALITY_LEVEL_2008:         "Windows Server 2008",
	DOMAIN_FUNCTIONALITY_LEVEL_2008_R2:      "Windows Server 2008 R2",
	DOMAIN_FUNCTIONALITY_LEVEL_2012:         "Windows Server 2012",
	DOMAIN_FUNCTIONALITY_LEVEL_2012_R2:      "Windows Server 2012 R2",
	DOMAIN_FUNCTIONALITY_LEVEL_2016:         "Windows Server 2016",
	DOMAIN_FUNCTIONALITY_LEVEL_2025:         "Windows Server 2025",
}

// String returns the string representation of the domain functionality level
func (v DomainFunctionalityLevel) String() string {
	if name, exists := DomainFunctionalityLevelToWindowsVersion[v]; exists {
		return fmt.Sprintf("%s", name)
	} else {
		return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}

// IsSupported returns true if the value is a domain functionality level this package
// recognises, that is one named by DomainFunctionalityLevelToWindowsVersion.
//
// The check consults that table rather than restating its contents as a chain of
// equality comparisons, which is how the two came to disagree: the comparison chain
// listed eight of the nine defined levels and omitted 2003 Interim.
//
// Returns:
//   - true if the level is one this package defines, false otherwise.
func (v DomainFunctionalityLevel) IsSupported() bool {
	_, exists := DomainFunctionalityLevelToWindowsVersion[v]

	return exists
}

// Equal returns true if the domain functionality level is equal to the specified level
func (v DomainFunctionalityLevel) Equal(other DomainFunctionalityLevel) bool {
	return v == other
}

// Less returns true if the domain functionality level is less than the specified level
func (v DomainFunctionalityLevel) Less(other DomainFunctionalityLevel) bool {
	return v < other
}
