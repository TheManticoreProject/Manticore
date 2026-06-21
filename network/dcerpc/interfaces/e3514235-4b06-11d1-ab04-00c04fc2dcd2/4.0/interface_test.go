package rpcinterface_e35142354b0611d1ab0400c04fc2dcd2_4_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the drsuapi interface
// (e3514235-4b06-11d1-ab04-00c04fc2dcd2 v4.0, [MS-DRSR]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "e3514235-4b06-11d1-ab04-00c04fc2dcd2" {
		t.Errorf("UUID = %s, want e3514235-4b06-11d1-ab04-00c04fc2dcd2", got)
	}
	if s.MajorVersion != 4 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 4.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses, so the
// two maps never drift, and that the on-the-wire opnum range (0..30) is fully covered.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 31 {
		t.Fatalf("OpnumToName has %d entries, want 31 (opnums 0..30)", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for op := uint16(0); op <= 30; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusString checks a couple of mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorDsDraAccessDenied); got != "ERROR_DS_DRA_ACCESS_DENIED" {
		t.Errorf("StatusString(0x2105) = %q, want ERROR_DS_DRA_ACCESS_DENIED", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}
