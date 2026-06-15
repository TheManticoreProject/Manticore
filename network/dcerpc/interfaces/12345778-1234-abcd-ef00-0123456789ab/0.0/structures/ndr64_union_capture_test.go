package structures

import (
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// policyInfoResponse mirrors the LsarQueryInformationPolicy reply: the [out] union as a
// [unique] pointer, then the NTSTATUS return value.
type policyInfoResponse struct {
	PolicyInformation *LSAPR_POLICY_INFORMATION `ndr:"unique"`
	Status            uint32                    `ndr:"retval"`
}

// TestNDR64_PolicyInformationUnion_FromCapture decodes LSAPR_POLICY_INFORMATION union
// responses captured from a live Windows Server 2016 LsarQueryInformationPolicy exchange
// (NDR64), confirming the NDR64 union decoder: an 8-octet [unique] referent, a 4-octet
// enum discriminant, and the arm aligned to 8.
func TestNDR64_PolicyInformationUnion_FromCapture(t *testing.T) {
	// Server-role arm (class 6): role = 3.
	roleStub, _ := hex.DecodeString("000002000000000006000000000000000300000000000000")
	var role policyInfoResponse
	if err := ndr.UnmarshalAs(roleStub, &role, ndr.NDR64); err != nil {
		t.Fatalf("decode role union: %v", err)
	}
	if role.PolicyInformation == nil || role.PolicyInformation.Class != PolicyLsaServerRoleInformation {
		t.Fatalf("role: Class = %v, want %d", role.PolicyInformation, PolicyLsaServerRoleInformation)
	}
	if got := role.PolicyInformation.PolicyServerRoleInfo.LsaServerRole; got != 3 {
		t.Errorf("role: LsaServerRole = %d, want 3", got)
	}

	// Account-domain arm (class 5): name + SID.
	domStub, _ := hex.DecodeString("000002000000000005000000000000001600180000000000000002000000000000000200000000000c0000000000000000000000000000000b0000000000000054004d0050002d0057002d003200300031003600300000000400000000000000010400000000000515000000fdca06a7cba8d22c3eae86ec00000000")
	var dom policyInfoResponse
	if err := ndr.UnmarshalAs(domStub, &dom, ndr.NDR64); err != nil {
		t.Fatalf("decode account-domain union: %v", err)
	}
	if dom.PolicyInformation == nil || dom.PolicyInformation.Class != PolicyAccountDomainInformation {
		t.Fatalf("account-domain: Class = %v, want %d", dom.PolicyInformation, PolicyAccountDomainInformation)
	}
	info := dom.PolicyInformation.PolicyAccountDomainInfo
	if name := info.DomainName.String(); name != "TMP-W-20160" {
		t.Errorf("account-domain: DomainName = %q, want %q", name, "TMP-W-20160")
	}
	if info.DomainSid == nil {
		t.Fatal("account-domain: DomainSid is nil")
	}
	if sid := info.DomainSid.String(); sid != "S-1-5-21-2802240253-752003275-3968249406" {
		t.Errorf("account-domain: DomainSid = %q, want %q", sid, "S-1-5-21-2802240253-752003275-3968249406")
	}
}
