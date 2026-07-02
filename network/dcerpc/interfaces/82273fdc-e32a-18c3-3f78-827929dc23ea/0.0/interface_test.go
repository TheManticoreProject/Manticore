package rpcinterface_82273fdce32a18c33f78827929dc23ea_0_0

import "testing"

// TestSyntaxID pins the abstract syntax: 82273fdc-e32a-18c3-3f78-827929dc23ea v0.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	u := s.UUID
	if u.A != 0x82273fdc || u.B != 0xe32a || u.C != 0x18c3 || u.D != 0x3f78 || u.E != 0x827929dc23ea {
		t.Errorf("UUID = %s, want 82273fdc-e32a-18c3-3f78-827929dc23ea", u.ToFormatD())
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\eventlog` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\eventlog`)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent, that the
// 23 on-the-wire opnums are present exactly once, and that the four "not used on the
// wire" opnums (19, 20, 21, 23) are absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	wireOpnums := []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 22, 24, 25, 26}
	if len(OpnumToName) != len(wireOpnums) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(wireOpnums))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	for _, op := range wireOpnums {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("on-the-wire opnum %d missing from OpnumToName", op)
		}
	}
	for _, gap := range []uint16{19, 20, 21, 23} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("opnum %d (%q) is NotUsedOnWire and must be absent", gap, name)
		}
	}
}

// TestStatusString spot-checks a few mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:            "ERROR_SUCCESS",
		ErrorAccessDenied:        "ERROR_ACCESS_DENIED",
		ErrorHandleEOF:           "ERROR_HANDLE_EOF",
		ErrorEventlogFileChanged: "ERROR_EVENTLOG_FILE_CHANGED",
		StatusBufferTooSmall:     "STATUS_BUFFER_TOO_SMALL",
		0xdeadbeef:               "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
}
