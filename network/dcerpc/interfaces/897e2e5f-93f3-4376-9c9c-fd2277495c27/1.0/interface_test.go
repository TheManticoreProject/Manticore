package rpcinterface_897e2e5f93f343769c9cfd2277495c27_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the FrsTransport interface
// (897e2e5f-93f3-4376-9c9c-fd2277495c27 v1.0, [MS-FRS2]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "897e2e5f-93f3-4376-9c9c-fd2277495c27" {
		t.Errorf("UUID = %s, want 897e2e5f-93f3-4376-9c9c-fd2277495c27", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// every on-the-wire opnum is covered (opnum 14 is NotUsedOnWire and must be absent).
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 17 {
		t.Fatalf("OpnumToName has %d entries, want 17 on-the-wire opnums", len(OpnumToName))
	}
	if _, ok := OpnumToName[14]; ok {
		t.Errorf("opnum 14 is NotUsedOnWire but present in OpnumToName")
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestStatusString checks a known mnemonic, an FRS-specific code, and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(FRS_ERROR_CONNECTION_INVALID); got != "FRS_ERROR_CONNECTION_INVALID" {
		t.Errorf("StatusString(0x2342) = %q, want FRS_ERROR_CONNECTION_INVALID", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}
