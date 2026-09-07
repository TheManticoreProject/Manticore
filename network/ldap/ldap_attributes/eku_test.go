package ldap_attributes_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

// Each OID is pinned against the value its name denotes in the Microsoft
// cryptography OID list and the PKIX arc. The file previously carried these
// unverified — its opening comment said so — and three were wrong.
//
// Src: https://mskb.pkisolutions.com/kb/287547
func TestEKUOIDs(t *testing.T) {
	testCases := []struct {
		name     string
		oid      string
		expected string
	}{
		// PKIX extended key purposes (RFC 5280 id-kp arc).
		{name: "EKU_SERVER_AUTHENTICATION", oid: ldap_attributes.EKU_SERVER_AUTHENTICATION, expected: "1.3.6.1.5.5.7.3.1"},
		{name: "EKU_CLIENT_AUTHENTICATION", oid: ldap_attributes.EKU_CLIENT_AUTHENTICATION, expected: "1.3.6.1.5.5.7.3.2"},
		{name: "EKU_CODE_SIGNING", oid: ldap_attributes.EKU_CODE_SIGNING, expected: "1.3.6.1.5.5.7.3.3"},
		{name: "EKU_EMAIL_PROTECTION", oid: ldap_attributes.EKU_EMAIL_PROTECTION, expected: "1.3.6.1.5.5.7.3.4"},
		{name: "EKU_IPSEC_END_SYSTEM", oid: ldap_attributes.EKU_IPSEC_END_SYSTEM, expected: "1.3.6.1.5.5.7.3.5"},
		{name: "EKU_IPSEC_TUNNEL", oid: ldap_attributes.EKU_IPSEC_TUNNEL, expected: "1.3.6.1.5.5.7.3.6"},
		{name: "EKU_IPSEC_USER", oid: ldap_attributes.EKU_IPSEC_USER, expected: "1.3.6.1.5.5.7.3.7"},
		{name: "EKU_TIME_STAMPING", oid: ldap_attributes.EKU_TIME_STAMPING, expected: "1.3.6.1.5.5.7.3.8"},
		{name: "EKU_OCSP_SIGNING", oid: ldap_attributes.EKU_OCSP_SIGNING, expected: "1.3.6.1.5.5.7.3.9"},
		{name: "EKU_ANY", oid: ldap_attributes.EKU_ANY, expected: "2.5.29.37.0"},
		{name: "EKU_KDC_AUTHENTICATION", oid: ldap_attributes.EKU_KDC_AUTHENTICATION, expected: "1.3.6.1.5.2.3.5"},

		// Microsoft OIDs.
		{name: "EKU_CERTIFICATE_REQUEST_AGENT (szOID_ENROLLMENT_AGENT)", oid: ldap_attributes.EKU_CERTIFICATE_REQUEST_AGENT, expected: "1.3.6.1.4.1.311.20.2.1"},
		{name: "EKU_SMART_CARD_LOGON (szOID_KP_SMARTCARD_LOGON)", oid: ldap_attributes.EKU_SMART_CARD_LOGON, expected: "1.3.6.1.4.1.311.20.2.2"},
		{name: "EKU_CA_EXCHANGE (szOID_KP_CA_EXCHANGE)", oid: ldap_attributes.EKU_CA_EXCHANGE, expected: "1.3.6.1.4.1.311.21.5"},
		{name: "EKU_KEY_RECOVERY_AGENT (szOID_KP_KEY_RECOVERY_AGENT)", oid: ldap_attributes.EKU_KEY_RECOVERY_AGENT, expected: "1.3.6.1.4.1.311.21.6"},
		{name: "EKU_DS_EMAIL_REPLICATION (szOID_DS_EMAIL_REPLICATION)", oid: ldap_attributes.EKU_DS_EMAIL_REPLICATION, expected: "1.3.6.1.4.1.311.21.19"},
		{name: "EKU_QUALIFIED_SUBORDINATION (szOID_KP_QUALIFIED_SUBORDINATION)", oid: ldap_attributes.EKU_QUALIFIED_SUBORDINATION, expected: "1.3.6.1.4.1.311.10.3.10"},
		{name: "EKU_DOCUMENT_SIGNING (szOID_KP_DOCUMENT_SIGNING)", oid: ldap_attributes.EKU_DOCUMENT_SIGNING, expected: "1.3.6.1.4.1.311.10.3.12"},
		{name: "EKU_LIFETIME_SIGNING (szOID_KP_LIFETIME_SIGNING)", oid: ldap_attributes.EKU_LIFETIME_SIGNING, expected: "1.3.6.1.4.1.311.10.3.13"},

		// The three that were wrong.
		{name: "EKU_ENCRYPTING_FILE_SYSTEM (szOID_EFS_CRYPTO)", oid: ldap_attributes.EKU_ENCRYPTING_FILE_SYSTEM, expected: "1.3.6.1.4.1.311.10.3.4"},
		{name: "EKU_FILE_RECOVERY (szOID_EFS_RECOVERY)", oid: ldap_attributes.EKU_FILE_RECOVERY, expected: "1.3.6.1.4.1.311.10.3.4.1"},
		{name: "EKU_KEY_PACK_LICENSES (szOID_LICENSES)", oid: ldap_attributes.EKU_KEY_PACK_LICENSES, expected: "1.3.6.1.4.1.311.10.6.1"},
		{name: "EKU_LICENSE_SERVER (szOID_LICENSE_SERVER)", oid: ldap_attributes.EKU_LICENSE_SERVER, expected: "1.3.6.1.4.1.311.10.6.2"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.oid != testCase.expected {
				t.Errorf("%s = %q, want %q", testCase.name, testCase.oid, testCase.expected)
			}
		})
	}
}

// EFS and file recovery are distinct OIDs one component apart, and conflating them is
// exactly what happened here: EKU_FILE_RECOVERY used to hold the EFS OID.
func TestEKUEFSAndFileRecoveryAreDistinct(t *testing.T) {
	if ldap_attributes.EKU_ENCRYPTING_FILE_SYSTEM == ldap_attributes.EKU_FILE_RECOVERY {
		t.Fatalf("EFS and file recovery share the OID %q", ldap_attributes.EKU_FILE_RECOVERY)
	}

	if ldap_attributes.EKU_FILE_RECOVERY != ldap_attributes.EKU_ENCRYPTING_FILE_SYSTEM+".1" {
		t.Errorf("file recovery = %q, want the child arc of EFS (%q.1)",
			ldap_attributes.EKU_FILE_RECOVERY, ldap_attributes.EKU_ENCRYPTING_FILE_SYSTEM)
	}
}

// No two constants may share an OID: a duplicate means one of them is mislabelled,
// which is how the two licensing OIDs came to be swapped.
func TestEKUOIDsAreUnique(t *testing.T) {
	oids := map[string]string{
		"EKU_CLIENT_AUTHENTICATION":     ldap_attributes.EKU_CLIENT_AUTHENTICATION,
		"EKU_SERVER_AUTHENTICATION":     ldap_attributes.EKU_SERVER_AUTHENTICATION,
		"EKU_CODE_SIGNING":              ldap_attributes.EKU_CODE_SIGNING,
		"EKU_EMAIL_PROTECTION":          ldap_attributes.EKU_EMAIL_PROTECTION,
		"EKU_TIME_STAMPING":             ldap_attributes.EKU_TIME_STAMPING,
		"EKU_OCSP_SIGNING":              ldap_attributes.EKU_OCSP_SIGNING,
		"EKU_IPSEC_END_SYSTEM":          ldap_attributes.EKU_IPSEC_END_SYSTEM,
		"EKU_IPSEC_TUNNEL":              ldap_attributes.EKU_IPSEC_TUNNEL,
		"EKU_IPSEC_USER":                ldap_attributes.EKU_IPSEC_USER,
		"EKU_ANY":                       ldap_attributes.EKU_ANY,
		"EKU_CERTIFICATE_REQUEST_AGENT": ldap_attributes.EKU_CERTIFICATE_REQUEST_AGENT,
		"EKU_SMART_CARD_LOGON":          ldap_attributes.EKU_SMART_CARD_LOGON,
		"EKU_DS_EMAIL_REPLICATION":      ldap_attributes.EKU_DS_EMAIL_REPLICATION,
		"EKU_KDC_AUTHENTICATION":        ldap_attributes.EKU_KDC_AUTHENTICATION,
		"EKU_ENCRYPTING_FILE_SYSTEM":    ldap_attributes.EKU_ENCRYPTING_FILE_SYSTEM,
		"EKU_FILE_RECOVERY":             ldap_attributes.EKU_FILE_RECOVERY,
		"EKU_QUALIFIED_SUBORDINATION":   ldap_attributes.EKU_QUALIFIED_SUBORDINATION,
		"EKU_KEY_RECOVERY_AGENT":        ldap_attributes.EKU_KEY_RECOVERY_AGENT,
		"EKU_CA_EXCHANGE":               ldap_attributes.EKU_CA_EXCHANGE,
		"EKU_LIFETIME_SIGNING":          ldap_attributes.EKU_LIFETIME_SIGNING,
		"EKU_DOCUMENT_SIGNING":          ldap_attributes.EKU_DOCUMENT_SIGNING,
		"EKU_KEY_PACK_LICENSES":         ldap_attributes.EKU_KEY_PACK_LICENSES,
		"EKU_LICENSE_SERVER":            ldap_attributes.EKU_LICENSE_SERVER,
	}

	seen := map[string]string{}
	for name, oid := range oids {
		if previous, ok := seen[oid]; ok {
			t.Errorf("%s and %s both hold %q", previous, name, oid)
		}
		seen[oid] = name
	}
}
