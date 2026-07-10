package ldap

import (
	"encoding/binary"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/winacl/securitydescriptor"
)

// impacketRBCDVector is a known-good msDS-AllowedToActOnBehalfOfOtherIdentity
// descriptor (captured from a reference implementation and confirmed accepted by
// a live Windows KDC for S4U2Proxy). It allows a single attacker SID.
const (
	impacketRBCDSID   = "S-1-5-21-2802240253-752003275-3968249406-194953"
	impacketRBCDBytes = "010004804000000000000000000000001400000004002c000100000000002400ff010f00010500000000000515000000fdca06a7cba8d22c3eae86ec89f9020001020000000000052000000020020000"
)

// TestBuildRBCDDescriptorGoldenVector checks BuildRBCDDescriptor reproduces the
// exact bytes a Windows KDC accepts: ACL_REVISION_DS (4) and the specific
// full-control mask 0x000F01FF, not GENERIC_ALL.
func TestBuildRBCDDescriptorGoldenVector(t *testing.T) {
	raw, err := BuildRBCDDescriptor(impacketRBCDSID)
	if err != nil {
		t.Fatalf("BuildRBCDDescriptor: %v", err)
	}
	if got := hex.EncodeToString(raw); got != impacketRBCDBytes {
		t.Errorf("descriptor bytes mismatch:\n got  %s\n want %s", got, impacketRBCDBytes)
	}
}

// TestBuildRBCDDescriptorRoundTrip builds the RBCD descriptor for a SID and
// confirms it parses back to exactly that SID.
func TestBuildRBCDDescriptorRoundTrip(t *testing.T) {
	sid := "S-1-5-21-1004336348-1177238915-682003330-1123"
	raw, err := BuildRBCDDescriptor(sid)
	if err != nil {
		t.Fatalf("BuildRBCDDescriptor: %v", err)
	}

	sids, err := ParseRBCDDescriptor(raw)
	if err != nil {
		t.Fatalf("ParseRBCDDescriptor: %v", err)
	}
	if want := []string{sid}; !reflect.DeepEqual(sids, want) {
		t.Errorf("parsed SIDs: got %v want %v", sids, want)
	}
}

// TestBuildRBCDDescriptorBytes checks the structural wire bytes of the descriptor
// so a regression in the marshal path is caught with fixed expectations:
// revision 1, SE_DACL_PRESENT, owner = BUILTIN\Administrators (S-1-5-32-544), a
// DACL with exactly one ACCESS_ALLOWED ACE granting the specific full-control
// mask 0x000F01FF, and the attacker SID as the trustee.
func TestBuildRBCDDescriptorBytes(t *testing.T) {
	sid := "S-1-5-21-1004336348-1177238915-682003330-1123"
	raw, err := BuildRBCDDescriptor(sid)
	if err != nil {
		t.Fatalf("BuildRBCDDescriptor: %v", err)
	}

	if raw[0] != 0x01 {
		t.Errorf("revision: got 0x%02x want 0x01", raw[0])
	}
	// Control flags (offset 2, LE uint16) must have SE_DACL_PRESENT (0x0004).
	control := binary.LittleEndian.Uint16(raw[2:4])
	if control&0x0004 == 0 {
		t.Errorf("DACL-present flag not set in control 0x%04x", control)
	}
	// DACL revision (at OffsetDacl = 20) must be ACL_REVISION_DS (4).
	if raw[20] != 4 {
		t.Errorf("DACL revision: got %d want 4 (ACL_REVISION_DS)", raw[20])
	}

	// Re-parse and assert the ACE grants the specific full-control mask.
	ntsd := securitydescriptor.NewSecurityDescriptor()
	if _, err := ntsd.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	owner := ntsd.GetOwner()
	if owner == nil || owner.SID.String() != "S-1-5-32-544" {
		t.Errorf("owner: got %v want BUILTIN\\Administrators (S-1-5-32-544)", owner)
	}
	dacl := ntsd.GetDacl()
	if dacl == nil || len(dacl.Entries) != 1 {
		t.Fatalf("expected exactly one DACL ACE, got %+v", dacl)
	}
	const fullControl = 0x000F01FF
	if dacl.Entries[0].Mask.RawValue != fullControl {
		t.Errorf("ACE mask: got 0x%08x want full-control 0x%08x", dacl.Entries[0].Mask.RawValue, fullControl)
	}
	if dacl.Entries[0].Identity.SID.String() != sid {
		t.Errorf("ACE trustee: got %s want %s", dacl.Entries[0].Identity.SID.String(), sid)
	}
}

// TestBuildRBCDDescriptorMultiSID confirms multiple allowed SIDs produce one ACE
// each, in order.
func TestBuildRBCDDescriptorMultiSID(t *testing.T) {
	a := "S-1-5-21-1004336348-1177238915-682003330-1123"
	b := "S-1-5-21-1004336348-1177238915-682003330-1124"
	raw, err := BuildRBCDDescriptor(a, b)
	if err != nil {
		t.Fatalf("BuildRBCDDescriptor: %v", err)
	}
	sids, err := ParseRBCDDescriptor(raw)
	if err != nil {
		t.Fatalf("ParseRBCDDescriptor: %v", err)
	}
	if want := []string{a, b}; !reflect.DeepEqual(sids, want) {
		t.Errorf("multi-SID: got %v want %v", sids, want)
	}
}

// TestBuildRBCDDescriptorErrors covers empty input and an invalid SID.
func TestBuildRBCDDescriptorErrors(t *testing.T) {
	if _, err := BuildRBCDDescriptor(); err == nil {
		t.Error("expected error with no SIDs")
	}
	if _, err := BuildRBCDDescriptor("not-a-sid"); err == nil {
		t.Error("expected error for invalid SID")
	}
}

// TestParseRBCDDescriptorEmpty confirms an empty attribute parses to no SIDs.
func TestParseRBCDDescriptorEmpty(t *testing.T) {
	sids, err := ParseRBCDDescriptor(nil)
	if err != nil {
		t.Fatalf("ParseRBCDDescriptor(nil): %v", err)
	}
	if sids != nil {
		t.Errorf("expected nil SIDs for empty descriptor, got %v", sids)
	}
}
