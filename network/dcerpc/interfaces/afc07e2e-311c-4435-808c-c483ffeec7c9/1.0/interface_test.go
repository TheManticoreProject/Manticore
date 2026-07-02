package rpcinterface_afc07e2e311c4435808cc483ffeec7c9_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the lsacap interface
// (afc07e2e-311c-4435-808c-c483ffeec7c9 v1.0, [MS-CAPR]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "afc07e2e-311c-4435-808c-c483ffeec7c9" {
		t.Errorf("UUID = %s, want afc07e2e-311c-4435-808c-c483ffeec7c9", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that the single on-the-wire opnum (0) is covered.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 1 {
		t.Fatalf("OpnumToName has %d entries, want 1 (opnum 0)", len(OpnumToName))
	}
	if OpnumToName[OpnumLsarGetAvailableCAPIDs] != "LsarGetAvailableCAPIDs" {
		t.Errorf("opnum 0 = %q, want LsarGetAvailableCAPIDs", OpnumToName[OpnumLsarGetAvailableCAPIDs])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestPipeName checks lsacap rides the shared LSA pipe.
func TestPipeName(t *testing.T) {
	if PipeName != `\lsarpc` {
		t.Errorf("PipeName = %q, want \\lsarpc", PipeName)
	}
}

// TestStatusString checks known mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "STATUS_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want STATUS_SUCCESS", got)
	}
	if got := StatusString(StatusAccessDenied); got != "STATUS_ACCESS_DENIED" {
		t.Errorf("StatusString(0xC0000022) = %q, want STATUS_ACCESS_DENIED", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}
