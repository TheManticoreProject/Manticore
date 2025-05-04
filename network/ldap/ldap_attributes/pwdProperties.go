package ldap_attributes

type PasswordProperties uint32

// PasswordProperties
// Src: https://learn.microsoft.com/en-us/windows/win32/api/ntsecapi/ns-ntsecapi-domain_password_information
const (
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_COMPLEX         PasswordProperties = 1
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_NO_ANON_CHANGE  PasswordProperties = 2
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_NO_CLEAR_CHANGE PasswordProperties = 4
	PASSWORD_PROPERTY_DOMAIN_LOCKOUT_ADMINS           PasswordProperties = 8
	PASSWORD_PROPERTY_DOMAIN_PASSWORD_STORE_CLEARTEXT PasswordProperties = 16
	PASSWORD_PROPERTY_DOMAIN_REFUSE_PASSWORD_CHANGE   PasswordProperties = 32
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

func (pwdProperties PasswordProperties) String() string {
	if _, ok := PasswordPropertiesMap[pwdProperties]; ok {
		return PasswordPropertiesMap[pwdProperties]
	}
	return ""
}

func (pwdProperties PasswordProperties) Description() string {
	if _, ok := PasswordPropertiesDescriptions[pwdProperties]; ok {
		return PasswordPropertiesDescriptions[pwdProperties]
	}
	return ""
}
