package rpcinterface_82ad4280036b11cf972c00aa006887b0_2_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the inetinfo interface
// (82ad4280-036b-11cf-972c-00aa006887b0 v2.0, [MS-IRP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "82ad4280-036b-11cf-972c-00aa006887b0" {
		t.Errorf("UUID = %s, want 82ad4280-036b-11cf-972c-00aa006887b0", got)
	}
	if s.MajorVersion != 2 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 2.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that all 16 on-the-wire opnums (0..15) are covered. Opnums 16 and 17
// (Opnum16NotUsedOnWire / Opnum17NotUsedOnWire) are intentionally absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 16 {
		t.Fatalf("OpnumToName has %d entries, want 16 (opnums 0..15)", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for op := uint16(0); op < 16; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	if _, ok := OpnumToName[16]; ok {
		t.Errorf("opnum 16 is not used on the wire and must be absent")
	}
}

// TestStatusString checks a known mnemonic and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(ErrorSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the [MS-IRP] 2.1.1 well-known endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\PIPE\inetinfo` {
		t.Errorf("PipeName = %q, want \\PIPE\\inetinfo", PipeName)
	}
}
