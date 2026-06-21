package rpcinterface_7c44d7d431d5424cbd5e2b3e1f323d22_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the dsaop interface
// (7c44d7d4-31d5-424c-bd5e-2b3e1f323d22 v1.0, [MS-DRSR]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "7c44d7d4-31d5-424c-bd5e-2b3e1f323d22" {
		t.Errorf("UUID = %s, want 7c44d7d4-31d5-424c-bd5e-2b3e1f323d22", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// both on-the-wire opnums (0..1) are covered.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 2 {
		t.Fatalf("OpnumToName has %d entries, want 2 (opnums 0..1)", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestStatusString checks a known mnemonic and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}
