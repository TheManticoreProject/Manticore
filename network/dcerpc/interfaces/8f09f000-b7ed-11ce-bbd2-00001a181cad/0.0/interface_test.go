package rpcinterface_8f09f000b7ed11cebbd200001a181cad_0_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the dimsvc interface
// (8f09f000-b7ed-11ce-bbd2-00001a181cad v0.0, [MS-RRASM]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "8f09f000-b7ed-11ce-bbd2-00001a181cad" {
		t.Errorf("UUID = %s, want 8f09f000-b7ed-11ce-bbd2-00001a181cad", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that all 53 on-the-wire opnums (0..52) are covered.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 53 {
		t.Fatalf("OpnumToName has %d entries, want 53 (opnums 0..52)", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for op := uint16(0); op < 53; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusString checks a known mnemonic and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(StatusNoMoreItems); got != "ERROR_NO_MORE_ITEMS" {
		t.Errorf("StatusString(0x103) = %q, want ERROR_NO_MORE_ITEMS", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the [MS-RRASM] 2.1 well-known endpoint shared with rasrpc.
func TestPipeName(t *testing.T) {
	if PipeName != `\ROUTER` {
		t.Errorf("PipeName = %q, want \\ROUTER", PipeName)
	}
}
