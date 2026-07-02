package rpcinterface_6b5bdd1e528c422caf8ca4079be4fe48_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the RemoteFW interface
// (6b5bdd1e-528c-422c-af8c-a4079be4fe48 v1.0, [MS-FASP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "6b5bdd1e-528c-422c-af8c-a4079be4fe48" {
		t.Errorf("UUID = %s, want 6b5bdd1e-528c-422c-af8c-a4079be4fe48", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that the anchor opnums resolve to the expected method names. RemoteFW exposes 94
// contiguous opnums (0..93); none are "not used on the wire".
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 94 {
		t.Fatalf("OpnumToName has %d entries, want 94", len(OpnumToName))
	}
	if OpnumToName[OpnumRRPC_FWOpenPolicyStore] != "RRPC_FWOpenPolicyStore" {
		t.Errorf("opnum 0 = %q, want RRPC_FWOpenPolicyStore", OpnumToName[OpnumRRPC_FWOpenPolicyStore])
	}
	if OpnumToName[OpnumRRPC_FWQueryFirewallRules2_33] != "RRPC_FWQueryFirewallRules2_33" {
		t.Errorf("opnum 93 = %q, want RRPC_FWQueryFirewallRules2_33", OpnumToName[OpnumRRPC_FWQueryFirewallRules2_33])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestOpnumContiguous checks the opnum space is exactly 0..93 with no gaps.
func TestOpnumContiguous(t *testing.T) {
	for op := uint16(0); op <= 93; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	if _, ok := OpnumToName[94]; ok {
		t.Errorf("opnum 94 should not exist")
	}
}

// TestStatusString checks known Win32 mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorAccessDenied); got != "ERROR_ACCESS_DENIED" {
		t.Errorf("StatusString(5) = %q, want ERROR_ACCESS_DENIED", got)
	}
	if got := StatusString(ErrorInvalidParameter); got != "ERROR_INVALID_PARAMETER" {
		t.Errorf("StatusString(0x57) = %q, want ERROR_INVALID_PARAMETER", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName documents that this interface has no named pipe: MS-FASP uses
// ncacn_ip_tcp with a dynamic endpoint, so PipeName is intentionally empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}
