package ldap_attributes

type SAMAccountType uint32

// sAMAccountType Values
// Src: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/e742be45-665d-4576-b872-0bc99d1e1fbe
const (
	SAM_DOMAIN_OBJECT             = 0x00000000
	SAM_GROUP_OBJECT              = 0x10000000
	SAM_NON_SECURITY_GROUP_OBJECT = 0x10000001
	SAM_ALIAS_OBJECT              = 0x20000000
	SAM_NON_SECURITY_ALIAS_OBJECT = 0x20000001
	SAM_USER_OBJECT               = 0x30000000
	SAM_MACHINE_ACCOUNT           = 0x30000001
	SAM_TRUST_ACCOUNT             = 0x30000002
	SAM_APP_BASIC_GROUP           = 0x40000000
	SAM_APP_QUERY_GROUP           = 0x40000001
)

var SAMAccountTypeMap = map[SAMAccountType]string{
	SAM_DOMAIN_OBJECT:             "DOMAIN_OBJECT",
	SAM_GROUP_OBJECT:              "GROUP_OBJECT",
	SAM_NON_SECURITY_GROUP_OBJECT: "NON_SECURITY_GROUP_OBJECT",
	SAM_ALIAS_OBJECT:              "ALIAS_OBJECT",
	SAM_NON_SECURITY_ALIAS_OBJECT: "NON_SECURITY_ALIAS_OBJECT",
	SAM_USER_OBJECT:               "USER_OBJECT",
	SAM_MACHINE_ACCOUNT:           "MACHINE_ACCOUNT",
	SAM_TRUST_ACCOUNT:             "TRUST_ACCOUNT",
	SAM_APP_BASIC_GROUP:           "APP_BASIC_GROUP",
	SAM_APP_QUERY_GROUP:           "APP_QUERY_GROUP",
}

func (sam SAMAccountType) String() string {
	if _, ok := SAMAccountTypeMap[sam]; ok {
		return SAMAccountTypeMap[sam]
	}
	return "UNKNOWN"
}
