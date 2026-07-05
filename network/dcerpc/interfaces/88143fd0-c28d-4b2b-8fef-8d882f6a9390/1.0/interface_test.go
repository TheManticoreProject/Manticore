package rpcinterface_88143fd0c28d4b2b8fef8d882f6a9390_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the TermSrvEnumeration interface
// (88143fd0-c28d-4b2b-8fef-8d882f6a9390 v1.0, [MS-TSTS]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "88143fd0-c28d-4b2b-8fef-8d882f6a9390" {
		t.Errorf("UUID = %s, want 88143fd0-c28d-4b2b-8fef-8d882f6a9390", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) == 0 {
		t.Fatal("OpnumToName is empty")
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, n := range OpnumToName {
		if NameToOpnum[n] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, n, NameToOpnum[n])
		}
	}
}

// TestStatusString checks the success mnemonic and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "S_OK" {
		t.Errorf("StatusString(StatusSuccess) = %q, want S_OK", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the [MS-TSTS] section 1.9 endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\LSM_API_service` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\LSM_API_service`)
	}
}
