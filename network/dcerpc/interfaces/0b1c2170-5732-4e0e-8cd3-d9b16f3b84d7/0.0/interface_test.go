package rpcinterface_0b1c217057324e0e8cd3d9b16f3b84d7_0_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the authzr interface
// (0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7 v0.0, [MS-RAA]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7" {
		t.Errorf("UUID = %s, want 0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// the anchor opnums resolve to the expected method names. authzr exposes 7 contiguous
// opnums (0..6); none are "not used on the wire".
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 7 {
		t.Fatalf("OpnumToName has %d entries, want 7", len(OpnumToName))
	}
	if OpnumToName[OpnumAuthzrFreeContext] != "AuthzrFreeContext" {
		t.Errorf("opnum 0 = %q, want AuthzrFreeContext", OpnumToName[OpnumAuthzrFreeContext])
	}
	if OpnumToName[OpnumAuthzrModifySids] != "AuthzrModifySids" {
		t.Errorf("opnum 6 = %q, want AuthzrModifySids", OpnumToName[OpnumAuthzrModifySids])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestOpnumContiguous checks the opnum space is exactly 0..6 with no gaps.
func TestOpnumContiguous(t *testing.T) {
	for op := uint16(0); op <= 6; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	if _, ok := OpnumToName[7]; ok {
		t.Errorf("opnum 7 should not exist")
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

// TestPipeName documents that this interface has no named pipe: [MS-RAA] 2.1 uses
// ncacn_ip_tcp with a dynamic endpoint, so PipeName is intentionally empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}
