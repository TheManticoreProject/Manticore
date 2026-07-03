package rpcinterface_811109bfa4e111d1ab5400a0c91e9b45_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the winsi2 interface
// (811109bf-a4e1-11d1-ab54-00a0c91e9b45 v1.0, [MS-RAIW]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "811109bf-a4e1-11d1-ab54-00a0c91e9b45" {
		t.Errorf("UUID = %s, want 811109bf-a4e1-11d1-ab54-00a0c91e9b45", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses. winsi2
// exposes 2 contiguous opnums (0..1).
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 2 {
		t.Fatalf("OpnumToName has %d entries, want 2", len(OpnumToName))
	}
	if OpnumToName[OpnumR_WinsTombstoneDbRecs] != "R_WinsTombstoneDbRecs" {
		t.Errorf("opnum 0 = %q, want R_WinsTombstoneDbRecs", OpnumToName[OpnumR_WinsTombstoneDbRecs])
	}
	if OpnumToName[OpnumR_WinsCheckAccess] != "R_WinsCheckAccess" {
		t.Errorf("opnum 1 = %q, want R_WinsCheckAccess", OpnumToName[OpnumR_WinsCheckAccess])
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
	if got := StatusString(ErrorWinsInternal); got != "ERROR_WINS_INTERNAL" {
		t.Errorf("StatusString(0xFA0) = %q, want ERROR_WINS_INTERNAL", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the shared WINS named pipe ([MS-RAIW] 2.1, Standards Assignments).
func TestPipeName(t *testing.T) {
	if PipeName != `\WinsPipe` {
		t.Errorf("PipeName = %q, want \\WinsPipe", PipeName)
	}
}
