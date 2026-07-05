package rpcinterface_5ca4a760ebb111cf861100a0245420ed_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the IcaApi interface
// (5ca4a760-ebb1-11cf-8611-00a0245420ed v1.0, [MS-TSTS]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "5ca4a760-ebb1-11cf-8611-00a0245420ed" {
		t.Errorf("UUID = %s, want 5ca4a760-ebb1-11cf-8611-00a0245420ed", got)
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
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(StatusSuccess) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the [MS-TSTS] section 1.9 endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\Ctx_WinStation_API_service` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\Ctx_WinStation_API_service`)
	}
}
