package rpcinterface_8fb6d884238811d08c3500c04fda2795_4_1

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the W32Time interface
// (8fb6d884-2388-11d0-8c35-00c04fda2795 v4.1, [MS-W32T]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "8fb6d884-2388-11d0-8c35-00c04fda2795" {
		t.Errorf("UUID = %s, want 8fb6d884-2388-11d0-8c35-00c04fda2795", got)
	}
	if s.MajorVersion != 4 || s.MinorVersion != 1 {
		t.Errorf("version = %d.%d, want 4.1", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover every opnum 0..7.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 8 {
		t.Fatalf("OpnumToName has %d entries, want 8", len(OpnumToName))
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

// TestStatusString checks a documented mnemonic, the success mnemonic, and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(StatusSuccess) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorAccessDenied); got != "ERROR_ACCESS_DENIED" {
		t.Errorf("StatusString(ErrorAccessDenied) = %q, want ERROR_ACCESS_DENIED", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the [MS-W32T] section 2.1 well-known endpoints.
func TestPipeName(t *testing.T) {
	if PipeName != `\W32TIME` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\W32TIME`)
	}
	if PipeNameAlt != `\W32TIME_ALT` {
		t.Errorf("PipeNameAlt = %q, want %q", PipeNameAlt, `\W32TIME_ALT`)
	}
}
