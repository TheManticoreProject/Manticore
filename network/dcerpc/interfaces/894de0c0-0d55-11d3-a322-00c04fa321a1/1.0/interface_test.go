package rpcinterface_894de0c00d5511d3a32200c04fa321a1_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the InitShutdown interface
// (894de0c0-0d55-11d3-a322-00c04fa321a1 v1.0, [MS-RSP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "894de0c0-0d55-11d3-a322-00c04fa321a1" {
		t.Errorf("UUID = %s, want 894de0c0-0d55-11d3-a322-00c04fa321a1", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover exactly the 3 on-the-wire opnums (0, 1, 2).
func TestOpnumNameRoundTrip(t *testing.T) {
	wire := []uint16{0, 1, 2}
	if len(OpnumToName) != len(wire) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(wire))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for _, op := range wire {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusString checks known mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:             "ERROR_SUCCESS",
		ErrorAccessDenied:         "ERROR_ACCESS_DENIED",
		ErrorShutdownInProgress:   "ERROR_SHUTDOWN_IN_PROGRESS",
		ErrorNoShutdownInProgress: "ERROR_NO_SHUTDOWN_IN_PROGRESS",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the [MS-RSP] 2.1 well-known endpoint for InitShutdown.
func TestPipeName(t *testing.T) {
	if PipeName != `\InitShutdown` {
		t.Errorf("PipeName = %q, want \\InitShutdown", PipeName)
	}
}
