package ldap_attributes_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

func TestUserAccountControlToString(t *testing.T) {
	testCases := []struct {
		name     string
		uac      ldap_attributes.UserAccountControl
		expected string
	}{
		{name: "No Flags", uac: ldap_attributes.UserAccountControl(0x00000000), expected: ""},
		{name: "Normal Account", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_NORMAL_ACCOUNT), expected: "NORMAL_ACCOUNT"},
		{name: "Script", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_SCRIPT), expected: "SCRIPT"},
		{name: "Account Disabled", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_ACCOUNT_DISABLED), expected: "ACCOUNT_DISABLED"},
		{name: "Home Directory Required", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_HOMEDIR_REQUIRED), expected: "HOMEDIR_REQUIRED"},
		{name: "Lockout", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_LOCKOUT), expected: "LOCKOUT"},
		{name: "Password Not Required", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_PASSWD_NOTREQD), expected: "PASSWD_NOTREQD"},
		{name: "Password Can't Change", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_PASSWD_CANT_CHANGE), expected: "PASSWD_CANT_CHANGE"},
		{name: "Encrypted Text Password Allowed", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_ENCRYPTED_TEXT_PWD_ALLOWED), expected: "ENCRYPTED_TEXT_PWD_ALLOWED"},
		{name: "Temporary Duplicate Account", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_TEMP_DUPLICATE_ACCOUNT), expected: "TEMP_DUPLICATE_ACCOUNT"},
		{name: "Normal Account", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_NORMAL_ACCOUNT), expected: "NORMAL_ACCOUNT"},
		{name: "Interdomain Trust Account", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_INTERDOMAIN_TRUST_ACCOUNT), expected: "INTERDOMAIN_TRUST_ACCOUNT"},
		{name: "Workstation Trust Account", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_WORKSTATION_TRUST_ACCOUNT), expected: "WORKSTATION_TRUST_ACCOUNT"},
		{name: "Server Trust Account", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_SERVER_TRUST_ACCOUNT), expected: "SERVER_TRUST_ACCOUNT"},
		{name: "Don't Expire Password", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_DONT_EXPIRE_PASSWORD), expected: "DONT_EXPIRE_PASSWORD"},
		{name: "MNS Logon Account", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_MNS_LOGON_ACCOUNT), expected: "MNS_LOGON_ACCOUNT"},
		{name: "Smartcard Required", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_SMARTCARD_REQUIRED), expected: "SMARTCARD_REQUIRED"},
		{name: "Trusted For Delegation", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_TRUSTED_FOR_DELEGATION), expected: "TRUSTED_FOR_DELEGATION"},
		{name: "Not Delegated", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_NOT_DELEGATED), expected: "NOT_DELEGATED"},
		{name: "Use DES Key Only", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_USE_DES_KEY_ONLY), expected: "USE_DES_KEY_ONLY"},
		{name: "Don't Require Pre-Auth", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_DONT_REQ_PREAUTH), expected: "DONT_REQ_PREAUTH"},
		{name: "Password Expired", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_PASSWORD_EXPIRED), expected: "PASSWORD_EXPIRED"},
		{name: "Trusted To Auth For Delegation", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_TRUSTED_TO_AUTH_FOR_DELEGATION), expected: "TRUSTED_TO_AUTH_FOR_DELEGATION"},
		{name: "Partial Secrets Account", uac: ldap_attributes.UserAccountControl(ldap_attributes.UAF_PARTIAL_SECRETS_ACCOUNT), expected: "PARTIAL_SECRETS_ACCOUNT"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.uac.String() != testCase.expected {
				t.Errorf("Expected %s, got %s", testCase.expected, testCase.uac.String())
			}
		})
	}
}

// Each constant is pinned against the literal value from [MS-ADTS] 2.2.16, not
// against UserAccountControlMap. The pre-existing tests above assert a constant
// against a map keyed by that same constant, so they hold for whatever value the
// constant happens to carry — which is why HOMEDIR_REQUIRED sitting on 0x00000004
// and PARTIAL_SECRETS_ACCOUNT on 0x08000000 went unnoticed.
//
// Src: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/dd302fd1-0aa7-406b-ad91-2a6b35738557
func TestUserAccountControlSpecValues(t *testing.T) {
	testCases := []struct {
		name     string
		flag     ldap_attributes.UserAccountControl
		expected uint32
	}{
		{name: "SCRIPT", flag: ldap_attributes.UAF_SCRIPT, expected: 0x00000001},
		{name: "ACCOUNT_DISABLED", flag: ldap_attributes.UAF_ACCOUNT_DISABLED, expected: 0x00000002},
		{name: "HOMEDIR_REQUIRED", flag: ldap_attributes.UAF_HOMEDIR_REQUIRED, expected: 0x00000008},
		{name: "LOCKOUT", flag: ldap_attributes.UAF_LOCKOUT, expected: 0x00000010},
		{name: "PASSWD_NOTREQD", flag: ldap_attributes.UAF_PASSWD_NOTREQD, expected: 0x00000020},
		{name: "PASSWD_CANT_CHANGE", flag: ldap_attributes.UAF_PASSWD_CANT_CHANGE, expected: 0x00000040},
		{name: "ENCRYPTED_TEXT_PWD_ALLOWED", flag: ldap_attributes.UAF_ENCRYPTED_TEXT_PWD_ALLOWED, expected: 0x00000080},
		{name: "TEMP_DUPLICATE_ACCOUNT", flag: ldap_attributes.UAF_TEMP_DUPLICATE_ACCOUNT, expected: 0x00000100},
		{name: "NORMAL_ACCOUNT", flag: ldap_attributes.UAF_NORMAL_ACCOUNT, expected: 0x00000200},
		{name: "INTERDOMAIN_TRUST_ACCOUNT", flag: ldap_attributes.UAF_INTERDOMAIN_TRUST_ACCOUNT, expected: 0x00000800},
		{name: "WORKSTATION_TRUST_ACCOUNT", flag: ldap_attributes.UAF_WORKSTATION_TRUST_ACCOUNT, expected: 0x00001000},
		{name: "SERVER_TRUST_ACCOUNT", flag: ldap_attributes.UAF_SERVER_TRUST_ACCOUNT, expected: 0x00002000},
		{name: "DONT_EXPIRE_PASSWORD", flag: ldap_attributes.UAF_DONT_EXPIRE_PASSWORD, expected: 0x00010000},
		{name: "MNS_LOGON_ACCOUNT", flag: ldap_attributes.UAF_MNS_LOGON_ACCOUNT, expected: 0x00020000},
		{name: "SMARTCARD_REQUIRED", flag: ldap_attributes.UAF_SMARTCARD_REQUIRED, expected: 0x00040000},
		{name: "TRUSTED_FOR_DELEGATION", flag: ldap_attributes.UAF_TRUSTED_FOR_DELEGATION, expected: 0x00080000},
		{name: "NOT_DELEGATED", flag: ldap_attributes.UAF_NOT_DELEGATED, expected: 0x00100000},
		{name: "USE_DES_KEY_ONLY", flag: ldap_attributes.UAF_USE_DES_KEY_ONLY, expected: 0x00200000},
		{name: "DONT_REQ_PREAUTH", flag: ldap_attributes.UAF_DONT_REQ_PREAUTH, expected: 0x00400000},
		{name: "PASSWORD_EXPIRED", flag: ldap_attributes.UAF_PASSWORD_EXPIRED, expected: 0x00800000},
		{name: "TRUSTED_TO_AUTH_FOR_DELEGATION", flag: ldap_attributes.UAF_TRUSTED_TO_AUTH_FOR_DELEGATION, expected: 0x01000000},
		{name: "NO_AUTH_DATA_REQUIRED", flag: ldap_attributes.UAF_NO_AUTH_DATA_REQUIRED, expected: 0x02000000},
		{name: "PARTIAL_SECRETS_ACCOUNT", flag: ldap_attributes.UAF_PARTIAL_SECRETS_ACCOUNT, expected: 0x04000000},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if uint32(testCase.flag) != testCase.expected {
				t.Errorf("%s = 0x%08x (%d), want 0x%08x (%d)",
					testCase.name, uint32(testCase.flag), testCase.flag, testCase.expected, testCase.expected)
			}
		})
	}
}

// The bits MS-ADTS marks "X: Unused. Must be zero and ignored." must carry a
// reserved name and must not appear in the map, so String() never reports a flag
// for a bit the specification does not assign.
func TestUserAccountControlReservedBits(t *testing.T) {
	reserved := map[string]ldap_attributes.UserAccountControl{
		"bit 2":  ldap_attributes.UAF_RESERVED_02,
		"bit 10": ldap_attributes.UAF_RESERVED_10,
		"bit 14": ldap_attributes.UAF_RESERVED_14,
		"bit 15": ldap_attributes.UAF_RESERVED_15,
		"bit 27": ldap_attributes.UAF_RESERVED_27,
		"bit 28": ldap_attributes.UAF_RESERVED_28,
		"bit 29": ldap_attributes.UAF_RESERVED_29,
		"bit 30": ldap_attributes.UAF_RESERVED_30,
		"bit 31": ldap_attributes.UAF_RESERVED_31,
	}

	expectedValues := map[string]uint32{
		"bit 2":  0x00000004,
		"bit 10": 0x00000400,
		"bit 14": 0x00004000,
		"bit 15": 0x00008000,
		"bit 27": 0x08000000,
		"bit 28": 0x10000000,
		"bit 29": 0x20000000,
		"bit 30": 0x40000000,
		"bit 31": 0x80000000,
	}

	for name, flag := range reserved {
		t.Run(name, func(t *testing.T) {
			if uint32(flag) != expectedValues[name] {
				t.Errorf("%s = 0x%08x, want 0x%08x", name, uint32(flag), expectedValues[name])
			}

			if _, ok := ldap_attributes.UserAccountControlMap[flag]; ok {
				t.Errorf("%s (0x%08x) is named in UserAccountControlMap, want absent", name, uint32(flag))
			}

			if got := flag.String(); got != "" {
				t.Errorf("%s.String() = %q, want an empty string", name, got)
			}
		})
	}
}

// The RODC filter of GetAllReadOnlyDomainControllers renders this constant into a
// bitwise-AND matching rule, so its decimal form is what goes on the wire.
func TestPartialSecretsAccountFilterValue(t *testing.T) {
	if got := uint32(ldap_attributes.UAF_PARTIAL_SECRETS_ACCOUNT); got != 67108864 {
		t.Errorf("PARTIAL_SECRETS_ACCOUNT = %d, want 67108864 — the LDAP filter would read "+
			"(userAccountControl:1.2.840.113556.1.4.803:=%d)", got, got)
	}
}

// An account carrying the real HOMEDIR_REQUIRED bit must report it, and one carrying
// the unassigned neighbouring bit must report nothing.
func TestHomedirRequiredBitIsReported(t *testing.T) {
	if got := ldap_attributes.UserAccountControl(0x00000008).String(); got != "HOMEDIR_REQUIRED" {
		t.Errorf("UserAccountControl(0x00000008).String() = %q, want %q", got, "HOMEDIR_REQUIRED")
	}

	if got := ldap_attributes.UserAccountControl(0x00000004).String(); got != "" {
		t.Errorf("UserAccountControl(0x00000004).String() = %q, want an empty string", got)
	}
}

// A real RODC computer account: WORKSTATION_TRUST_ACCOUNT plus PARTIAL_SECRETS_ACCOUNT,
// which MS-ADTS requires to be set together.
func TestReadOnlyDomainControllerAccountFlags(t *testing.T) {
	rodc := ldap_attributes.UserAccountControl(0x04000000 | 0x00001000)

	if got, want := rodc.String(), "PARTIAL_SECRETS_ACCOUNT|WORKSTATION_TRUST_ACCOUNT"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
