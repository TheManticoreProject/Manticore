package rpcinterface_1a927394352e4553ae3f7cf4aafca620_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the WdsRpcInterface interface
// (1a927394-352e-4553-ae3f-7cf4aafca620 v1.0, [MS-WDSC]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "1a927394-352e-4553-ae3f-7cf4aafca620" {
		t.Errorf("UUID = %s, want 1a927394-352e-4553-ae3f-7cf4aafca620", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover the single opnum 0.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 1 {
		t.Fatalf("OpnumToName has %d entries, want 1", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	if OpnumToName[OpnumWdsRpcMessage] != "WdsRpcMessage" {
		t.Errorf("OpnumToName[%d] = %q, want WdsRpcMessage", OpnumWdsRpcMessage, OpnumToName[OpnumWdsRpcMessage])
	}
	for op, n := range OpnumToName {
		if NameToOpnum[n] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, n, NameToOpnum[n])
		}
	}
}

// TestStatusString checks the success mnemonic, a documented Win32 mnemonic, and the hex
// fallback for an unknown code.
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

// TestPipeName pins the descriptor-uniformity endpoint name. WdsRpcInterface is actually
// an ncacn_ip_tcp dynamic-endpoint interface ([MS-WDSC] 2.1); the constant is retained
// only for uniformity across interface descriptors.
func TestPipeName(t *testing.T) {
	if PipeName != `\WdsRpcInterface` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\WdsRpcInterface`)
	}
}
