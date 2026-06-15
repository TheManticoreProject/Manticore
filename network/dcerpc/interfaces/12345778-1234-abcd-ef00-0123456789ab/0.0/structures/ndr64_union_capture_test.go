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

	// DNS-domain arm (class 12): three counted strings, a 16-octet GUID, and a SID.
	dnsStub, _ := hex.DecodeString("00000200000000000c00000000000000160018000000000000000200000000002000220000000000000002000000000020002200000000000000020000000000480250a6c11c894eb2bfebe99677d08400000200000000000c0000000000000000000000000000000b0000000000000054004d0050002d0057002d0032003000310036003000000011000000000000000000000000000000100000000000000054004d0050002d0057002d0032003000310036002e006c006f00630061006c0011000000000000000000000000000000100000000000000054004d0050002d0057002d0032003000310036002e006c006f00630061006c000400000000000000010400000000000515000000fdca06a7cba8d22c3eae86ec00000000")
	var dns policyInfoResponse
	if err := ndr.UnmarshalAs(dnsStub, &dns, ndr.NDR64); err != nil {
		t.Fatalf("decode dns-domain union: %v", err)
	}
	if dns.PolicyInformation == nil || dns.PolicyInformation.Class != PolicyDnsDomainInformation {
		t.Fatalf("dns-domain: Class = %v, want %d", dns.PolicyInformation, PolicyDnsDomainInformation)
	}
	d := dns.PolicyInformation.PolicyDnsDomainInfo
	if d.Name.String() != "TMP-W-20160" || d.DnsDomainName.String() != "TMP-W-2016.local" || d.DnsForestName.String() != "TMP-W-2016.local" {
		t.Errorf("dns-domain: names = %q / %q / %q", d.Name.String(), d.DnsDomainName.String(), d.DnsForestName.String())
	}
	if g := d.DomainGuid.String(); g != "a6500248-1cc1-4e89-b2bf-ebe99677d084" {
		t.Errorf("dns-domain: DomainGuid = %q, want %q", g, "a6500248-1cc1-4e89-b2bf-ebe99677d084")
	}
	if d.Sid == nil || d.Sid.String() != "S-1-5-21-2802240253-752003275-3968249406" {
		t.Errorf("dns-domain: Sid = %v", d.Sid)
	}
}
