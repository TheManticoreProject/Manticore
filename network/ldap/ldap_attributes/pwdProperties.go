package ldap_attributes

import (
	"sort"
	"strings"
)

type PasswordProperties uint32

// PasswordProperties
// Src: https://learn.microsoft.com/en-us/windows/win32/api/ntsecapi/ns-ntsecapi-domain_password_information
const (
	// DOMAIN_PASSWORD_COMPLEX (0x00000001)
	// The password must have a mix of at least two of the following types of characters:
	// - Uppercase characters
	// - Lowercase characters
	// - Numerals
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_COMPLEX PasswordProperties = 0x00000001

	// DOMAIN_PASSWORD_NO_ANON_CHANGE (0x00000002)
	// The password cannot be changed without logging on. Otherwise, if your password has expired,
	// you can change your password and then log on.
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_NO_ANON_CHANGE PasswordProperties = 0x00000002

	// DOMAIN_PASSWORD_NO_CLEAR_CHANGE (0x00000004)
	// Forces the client to use a protocol that does not allow the domain controller to get the plaintext password.
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_NO_CLEAR_CHANGE PasswordProperties = 0x00000004

	// DOMAIN_LOCKOUT_ADMINS (0x00000008)
	// Allows the built-in administrator account to be locked out from network logons.
	PASSWORD_PROPERTY_DOMAIN_LOCKOUT_ADMINS PasswordProperties = 0x00000008

	// DOMAIN_PASSWORD_STORE_CLEARTEXT (0x00000010)
	// The directory service is storing a plaintext password for all users instead of a hash function of the password.
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_STORE_CLEARTEXT PasswordProperties = 0x00000010

	// DOMAIN_REFUSE_PASSWORD_CHANGE (0x00000020)
	// Removes the requirement that the machine account password be automatically changed every week.
	// This value should not be used as it can weaken security.
	PASSWORD_PROPERTY_DOMAIN_REFUSE_PASSWORD_CHANGE PasswordProperties = 0x00000020
)

var PasswordPropertiesMap = map[PasswordProperties]string{
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_COMPLEX:         "DOMAIN_PASSWORD_COMPLEX",
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_NO_ANON_CHANGE:  "DOMAIN_PASSWORD_NO_ANON_CHANGE",
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_NO_CLEAR_CHANGE: "DOMAIN_PASSWORD_NO_CLEAR_CHANGE",
	PASSWORD_PROPERTY_DOMAIN_LOCKOUT_ADMINS:           "DOMAIN_LOCKOUT_ADMINS",
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_STORE_CLEARTEXT: "DOMAIN_PASSWORD_STORE_CLEARTEXT",
	PASSWORD_PROPERTY_DOMAIN_REFUSE_PASSWORD_CHANGE:   "DOMAIN_REFUSE_PASSWORD_CHANGE",
}

var PasswordPropertiesDescriptions = map[PasswordProperties]string{
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_COMPLEX:         "The password must have a mix of at least two of the following types of characters: Uppercase characters, Lowercase characters, Numerals.",
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_NO_ANON_CHANGE:  "The password cannot be changed without logging on. Otherwise, if your password has expired, you can change your password and then log on.",
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_NO_CLEAR_CHANGE: "Forces the client to use a protocol that does not allow the domain controller to get the plaintext password.",
	PASSWORD_PROPERTY_DOMAIN_LOCKOUT_ADMINS:           "Allows the built-in administrator account to be locked out from network logons.",
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_STORE_CLEARTEXT: "The directory service is storing a plaintext password for all users instead of a hash function of the password.",
	PASSWORD_PROPERTY_DOMAIN_REFUSE_PASSWORD_CHANGE:   "Removes the requirement that the machine account password be automatically changed every week. This value should not be used as it can weaken security.",
}

// String returns the names of the flags set in the value, sorted alphabetically and
// separated by a pipe ("|").
//
// pwdProperties is a bit field and a domain policy routinely sets several flags at
// once, so each flag is tested with a bitwise AND rather than the whole value being
// looked up as a single key. A value with no known flag set returns an empty string.
//
// Returns:
//   - A string containing the names of the set flags, separated by a pipe ("|").
func (pwdProperties PasswordProperties) String() string {
	names := []string{}
	for flag, name := range PasswordPropertiesMap {
		if pwdProperties&flag != 0 {
			names = append(names, name)
		}
	}

	sort.Strings(names)

	return strings.Join(names, "|")
}

// Description returns the descriptions of the flags set in the value, one per line.
//
// As with String, the value is a bit field, so every set flag is described rather
// than the whole value being looked up as a single key.
//
// Returns:
//   - A string containing the descriptions of the set flags, separated by newlines.
func (pwdProperties PasswordProperties) Description() string {
	descriptions := []string{}
	for flag, description := range PasswordPropertiesDescriptions {
		if pwdProperties&flag != 0 {
			descriptions = append(descriptions, description)
		}
	}

	sort.Strings(descriptions)

	return strings.Join(descriptions, "\n")
}

// GetFlags returns the flags set in the value, sorted in ascending order.
//
// Returns:
//   - A slice of PasswordProperties values representing the set flags.
func (pwdProperties PasswordProperties) GetFlags() []PasswordProperties {
	flags := []PasswordProperties{}
	for flag := range PasswordPropertiesMap {
		if pwdProperties&flag != 0 {
			flags = append(flags, flag)
		}
	}

	sort.Slice(flags, func(i, j int) bool {
		return flags[i] < flags[j]
	})

	return flags
}
