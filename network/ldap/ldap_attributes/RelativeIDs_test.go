package ldap_attributes_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap/ldap_attributes"
)

// The two lists encode whether a RID's well-known SID is BUILTIN (S-1-5-32-<rid>) or
// domain-relative (S-1-5-<domain>-<rid>), and FindObjectSIDByRID builds the BUILTIN
// form for every LocalRIDs member. A domain-relative RID in LocalRIDs is therefore
// looked up as a SID that cannot exist.
//
// Src: https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-identifiers
func TestLocalRIDsContainsOnlyBuiltinRIDs(t *testing.T) {
	// RIDs Microsoft documents as S-1-5-<domain>-<rid> rather than S-1-5-32-<rid>.
	domainScoped := map[int]string{
		553: "RAS and IAS Servers",
		571: "Allowed RODC Password Replication Group",
		572: "Denied RODC Password Replication Group",
	}

	for _, rid := range ldap_attributes.LocalRIDs {
		if name, ok := domainScoped[rid]; ok {
			t.Errorf("LocalRIDs contains RID %d (0x%03X) %q, whose SID is S-1-5-<domain>-%d, not S-1-5-32-%d",
				rid, rid, name, rid, rid)
		}
	}
}

// The domain-relative RIDs have to be reachable through DomainRIDs so the domain form
// of the SID is built for them.
func TestDomainRIDsContainsTheDomainScopedRIDs(t *testing.T) {
	expected := map[int]string{
		ldap_attributes.RID_DOMAIN_ALIAS_RAS_SERVERS:                       "RAS and IAS Servers",
		ldap_attributes.RID_DOMAIN_GROUP_ALLOWED_RODC_PASSWORD_REPLICATION: "Allowed RODC Password Replication Group",
		ldap_attributes.RID_DOMAIN_GROUP_DENIED_RODC_PASSWORD_REPLICATION:  "Denied RODC Password Replication Group",
	}

	present := map[int]bool{}
	for _, rid := range ldap_attributes.DomainRIDs {
		present[rid] = true
	}

	for rid, name := range expected {
		if !present[rid] {
			t.Errorf("DomainRIDs is missing RID %d (0x%03X) %q", rid, rid, name)
		}
	}
}

// No RID may appear in both lists: LocalRIDs is scanned first, so a duplicate makes
// the domain form unreachable.
func TestRIDListsDoNotOverlap(t *testing.T) {
	inLocal := map[int]bool{}
	for _, rid := range ldap_attributes.LocalRIDs {
		inLocal[rid] = true
	}

	for _, rid := range ldap_attributes.DomainRIDs {
		if inLocal[rid] {
			t.Errorf("RID %d (0x%03X) is in both DomainRIDs and LocalRIDs", rid, rid)
		}
	}
}

// Neither list may contain the same RID twice under two names.
func TestRIDListsHaveNoInternalDuplicates(t *testing.T) {
	for _, testCase := range []struct {
		name string
		rids []int
	}{
		{name: "DomainRIDs", rids: ldap_attributes.DomainRIDs},
		{name: "LocalRIDs", rids: ldap_attributes.LocalRIDs},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			seen := map[int]bool{}
			for _, rid := range testCase.rids {
				if seen[rid] {
					t.Errorf("%s contains RID %d (0x%03X) more than once", testCase.name, rid, rid)
				}
				seen[rid] = true
			}
		})
	}
}

// The values of the RIDs this fix moved, pinned against Microsoft's table.
func TestDomainScopedRIDValues(t *testing.T) {
	testCases := []struct {
		name     string
		rid      int
		expected int
	}{
		{name: "RAS and IAS Servers", rid: ldap_attributes.RID_DOMAIN_ALIAS_RAS_SERVERS, expected: 553},
		{name: "Allowed RODC Password Replication", rid: ldap_attributes.RID_DOMAIN_GROUP_ALLOWED_RODC_PASSWORD_REPLICATION, expected: 571},
		{name: "Denied RODC Password Replication", rid: ldap_attributes.RID_DOMAIN_GROUP_DENIED_RODC_PASSWORD_REPLICATION, expected: 572},
		{name: "Certificate Service DCOM Access (BUILTIN)", rid: ldap_attributes.RID_LOCAL_CERTSVC_DCOM_ACCESS_GROUP, expected: 574},
		{name: "Event Log Readers (BUILTIN)", rid: ldap_attributes.RID_LOCAL_EVENT_LOG_READERS_GROUP, expected: 573},
		{name: "Cryptographic Operators (BUILTIN)", rid: ldap_attributes.RID_LOCAL_CRYPTO_OPERATORS, expected: 569},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.rid != testCase.expected {
				t.Errorf("%s = %d, want %d", testCase.name, testCase.rid, testCase.expected)
			}
		})
	}
}
