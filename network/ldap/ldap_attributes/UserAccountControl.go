package ldap_attributes

import (
	"sort"
	"strings"
)

type UserAccountControl uint32

// UserAccountControl
// Src: https://learn.microsoft.com/en-us/troubleshoot/windows-server/active-directory/useraccountcontrol-manipulate-account-properties
const (
	UAF_SCRIPT                         UserAccountControl = 1       // 1
	UAF_ACCOUNT_DISABLED               UserAccountControl = 1 << 1  // 2
	UAF_HOMEDIR_REQUIRED               UserAccountControl = 1 << 2  // 4
	UAF_RESERVED_03                    UserAccountControl = 1 << 3  // 8
	UAF_LOCKOUT                        UserAccountControl = 1 << 4  // 16
	UAF_PASSWD_NOTREQD                 UserAccountControl = 1 << 5  // 32
	UAF_PASSWD_CANT_CHANGE             UserAccountControl = 1 << 6  // 64
	UAF_ENCRYPTED_TEXT_PWD_ALLOWED     UserAccountControl = 1 << 7  // 128
	UAF_TEMP_DUPLICATE_ACCOUNT         UserAccountControl = 1 << 8  // 256
	UAF_NORMAL_ACCOUNT                 UserAccountControl = 1 << 9  // 512
	UAF_RESERVED_10                    UserAccountControl = 1 << 10 // 1024
	UAF_INTERDOMAIN_TRUST_ACCOUNT      UserAccountControl = 1 << 11 // 2048
	UAF_WORKSTATION_TRUST_ACCOUNT      UserAccountControl = 1 << 12 // 4096
	UAF_SERVER_TRUST_ACCOUNT           UserAccountControl = 1 << 13 // 8192
	UAF_RESERVED_14                    UserAccountControl = 1 << 14 // 16384
	UAF_RESERVED_15                    UserAccountControl = 1 << 15 // 32768
	UAF_DONT_EXPIRE_PASSWORD           UserAccountControl = 1 << 16 // 65536
	UAF_MNS_LOGON_ACCOUNT              UserAccountControl = 1 << 17 // 131072
	UAF_SMARTCARD_REQUIRED             UserAccountControl = 1 << 18 // 262144
	UAF_TRUSTED_FOR_DELEGATION         UserAccountControl = 1 << 19 // 524288
	UAF_NOT_DELEGATED                  UserAccountControl = 1 << 20 // 1048576
	UAF_USE_DES_KEY_ONLY               UserAccountControl = 1 << 21 // 2097152
	UAF_DONT_REQ_PREAUTH               UserAccountControl = 1 << 22 // 4194304
	UAF_PASSWORD_EXPIRED               UserAccountControl = 1 << 23 // 8388608
	UAF_TRUSTED_TO_AUTH_FOR_DELEGATION UserAccountControl = 1 << 24 // 16777216
	UAF_RESERVED_25                    UserAccountControl = 1 << 25 // 33554432
	UAF_RESERVED_26                    UserAccountControl = 1 << 26 // 67108864
	UAF_PARTIAL_SECRETS_ACCOUNT        UserAccountControl = 1 << 27 // 134217728
	UAF_RESERVED_28                    UserAccountControl = 1 << 28 // 268435456
	UAF_RESERVED_29                    UserAccountControl = 1 << 29 // 536870912
	UAF_RESERVED_30                    UserAccountControl = 1 << 30 // 1073741824
	UAF_RESERVED_31                    UserAccountControl = 1 << 31 // 2147483648
)

var UserAccountControlMap = map[UserAccountControl]string{
	UAF_SCRIPT:                         "SCRIPT",
	UAF_ACCOUNT_DISABLED:               "ACCOUNT_DISABLED",
	UAF_HOMEDIR_REQUIRED:               "HOMEDIR_REQUIRED",
	UAF_LOCKOUT:                        "LOCKOUT",
	UAF_PASSWD_NOTREQD:                 "PASSWD_NOTREQD",
	UAF_PASSWD_CANT_CHANGE:             "PASSWD_CANT_CHANGE",
	UAF_ENCRYPTED_TEXT_PWD_ALLOWED:     "ENCRYPTED_TEXT_PWD_ALLOWED",
	UAF_TEMP_DUPLICATE_ACCOUNT:         "TEMP_DUPLICATE_ACCOUNT",
	UAF_NORMAL_ACCOUNT:                 "NORMAL_ACCOUNT",
	UAF_INTERDOMAIN_TRUST_ACCOUNT:      "INTERDOMAIN_TRUST_ACCOUNT",
	UAF_WORKSTATION_TRUST_ACCOUNT:      "WORKSTATION_TRUST_ACCOUNT",
	UAF_SERVER_TRUST_ACCOUNT:           "SERVER_TRUST_ACCOUNT",
	UAF_DONT_EXPIRE_PASSWORD:           "DONT_EXPIRE_PASSWORD",
	UAF_MNS_LOGON_ACCOUNT:              "MNS_LOGON_ACCOUNT",
	UAF_SMARTCARD_REQUIRED:             "SMARTCARD_REQUIRED",
	UAF_TRUSTED_FOR_DELEGATION:         "TRUSTED_FOR_DELEGATION",
	UAF_NOT_DELEGATED:                  "NOT_DELEGATED",
	UAF_USE_DES_KEY_ONLY:               "USE_DES_KEY_ONLY",
	UAF_DONT_REQ_PREAUTH:               "DONT_REQ_PREAUTH",
	UAF_PASSWORD_EXPIRED:               "PASSWORD_EXPIRED",
	UAF_TRUSTED_TO_AUTH_FOR_DELEGATION: "TRUSTED_TO_AUTH_FOR_DELEGATION",
	UAF_PARTIAL_SECRETS_ACCOUNT:        "PARTIAL_SECRETS_ACCOUNT",
}

// String returns a string representation of the UserAccountControl flags.
//
// The function iterates over the UserAccountControlMap to check which flags are set in the UserAccountControl value.
// It collects the string representations of the set flags, sorts them alphabetically, and joins them with a pipe ("|") separator.
//
// Returns:
//   - A string containing the names of the set flags, separated by a pipe ("|").
//
// Example usage:
//
//	uac := ldap_attributes.UserAccountControl(0x00000010 | 0x00000020)
//	fmt.Println(uac.String()) // Output: "LOCKOUT|PASSWD_NOTREQD"
//
// This function is useful for debugging and logging purposes, allowing a human-readable representation of the UserAccountControl flags.
func (uac UserAccountControl) String() string {
	flagsString := []string{}
	for flag, val := range UserAccountControlMap {
		if uac&flag != 0 {
			flagsString = append(flagsString, val)
		}
	}

	sort.Strings(flagsString)

	return strings.Join(flagsString, "|")
}

// GetFlags returns a slice of UserAccountControl flags that are set in the UserAccountControl value.
//
// The function iterates over the UserAccountControlMap to check which flags are set in the UserAccountControl value.
// It collects the flags that are set and returns them as a slice of UserAccountControl values, sorted in ascending order.
//
// Returns:
//   - A slice of UserAccountControl values representing the set flags, sorted in ascending order.
//
// Example usage:
//
//	uac := ldap_attributes.UserAccountControl(0x00000010 | 0x00000020)
//	flags := uac.GetFlags()
//	for _, flag := range flags {
//	    fmt.Println(flag)
//	}
//
// This function is useful for obtaining a list of individual flags set in the UserAccountControl value, which can be
// used for further processing or analysis.
func (uac UserAccountControl) GetFlags() []UserAccountControl {
	flags := []UserAccountControl{}
	for uaf := range UserAccountControlMap {
		if uac&uaf != 0 {
			flags = append(flags, uaf)
		}
	}

	sort.Slice(flags, func(i, j int) bool {
		return flags[i] < flags[j]
	})

	return flags
}
