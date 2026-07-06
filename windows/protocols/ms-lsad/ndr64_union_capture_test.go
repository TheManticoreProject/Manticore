package mslsad

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// policyInfoResponse mirrors the LsarQueryInformationPolicy reply: the [out] union as a
// [unique] pointer, then the NTSTATUS return value.
type policyInfoResponse struct {
	PolicyInformation *LSAPR_POLICY_INFORMATION `ndr:"unique"`
	Status            uint32                    `ndr:"retval"`
}

// roundTripNDR64 marshals a policy-information union as the LsarQueryInformationPolicy
// [out] parameter under NDR64 and unmarshals it back, returning the decoded union.
func roundTripNDR64(t *testing.T, in *LSAPR_POLICY_INFORMATION) *LSAPR_POLICY_INFORMATION {
	t.Helper()
	wire, err := ndr.MarshalAs(&policyInfoResponse{PolicyInformation: in}, ndr.NDR64)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out policyInfoResponse
	if err := ndr.UnmarshalAs(wire, &out, ndr.NDR64); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.PolicyInformation == nil {
		t.Fatal("decoded PolicyInformation is nil")
	}
	return out.PolicyInformation
}

// TestNDR64_PolicyInformationUnion round-trips the LSAPR_POLICY_INFORMATION union under
// NDR64 across representative arms, exercising the 8-octet [unique] referent, the 4-octet
// enum discriminant, the 8-aligned arm, and the RPC_UNICODE_STRING / RPC_SID / GUID
// members. Values are synthetic; the on-wire framing was verified against a live capture
// during development (issue #596).
func TestNDR64_PolicyInformationUnion(t *testing.T) {
	sid, err := msdtyp.ParseSID("S-1-5-21-1-2-3")
	if err != nil {
		t.Fatalf("ParseSID: %v", err)
	}
	g := msdtyp.GUID{Data1: 0x11223344, Data2: 0x5566, Data3: 0x7788, Data4: [8]byte{0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00}}

	// Server-role arm (class 6): a 4-octet enum within the arm.
	role := roundTripNDR64(t, &LSAPR_POLICY_INFORMATION{
		Class:                PolicyLsaServerRoleInformation,
		PolicyServerRoleInfo: POLICY_LSA_SERVER_ROLE_INFO{LsaServerRole: PolicyServerRolePrimary},
	})
	if role.Class != PolicyLsaServerRoleInformation || role.PolicyServerRoleInfo.LsaServerRole != PolicyServerRolePrimary {
		t.Errorf("server-role arm: got %+v", role)
	}

	// Account-domain arm (class 5): counted string + [unique] SID.
	acct := roundTripNDR64(t, &LSAPR_POLICY_INFORMATION{
		Class:                   PolicyAccountDomainInformation,
		PolicyAccountDomainInfo: LSAPR_POLICY_ACCOUNT_DOM_INFO{DomainName: msdtyp.NewUnicodeString("TESTDOM"), DomainSid: &sid},
	})
	if acct.Class != PolicyAccountDomainInformation {
		t.Fatalf("account-domain class = %d", acct.Class)
	}
	if got := acct.PolicyAccountDomainInfo.DomainName.String(); got != "TESTDOM" {
		t.Errorf("account-domain name = %q, want TESTDOM", got)
	}
	if acct.PolicyAccountDomainInfo.DomainSid == nil || acct.PolicyAccountDomainInfo.DomainSid.String() != "S-1-5-21-1-2-3" {
		t.Errorf("account-domain SID = %v", acct.PolicyAccountDomainInfo.DomainSid)
	}

	// DNS-domain arm (class 12): three counted strings, a 16-octet GUID, and a SID.
	dns := roundTripNDR64(t, &LSAPR_POLICY_INFORMATION{
		Class: PolicyDnsDomainInformation,
		PolicyDnsDomainInfo: LSAPR_POLICY_DNS_DOMAIN_INFO{
			Name:          msdtyp.NewUnicodeString("TESTDOM"),
			DnsDomainName: msdtyp.NewUnicodeString("test.example"),
			DnsForestName: msdtyp.NewUnicodeString("test.example"),
			DomainGuid:    g,
			Sid:           &sid,
		},
	})
	d := dns.PolicyDnsDomainInfo
	if d.Name.String() != "TESTDOM" || d.DnsDomainName.String() != "test.example" || d.DnsForestName.String() != "test.example" {
		t.Errorf("dns-domain names = %q / %q / %q", d.Name.String(), d.DnsDomainName.String(), d.DnsForestName.String())
	}
	if d.DomainGuid != g {
		t.Errorf("dns-domain GUID = %v, want %v", d.DomainGuid, g)
	}
	if d.Sid == nil || d.Sid.String() != "S-1-5-21-1-2-3" {
		t.Errorf("dns-domain SID = %v", d.Sid)
	}
}
