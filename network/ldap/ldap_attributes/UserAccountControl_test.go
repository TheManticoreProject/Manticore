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
