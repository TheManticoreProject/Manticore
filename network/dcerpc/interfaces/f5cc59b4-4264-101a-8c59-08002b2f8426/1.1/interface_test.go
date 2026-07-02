package rpcinterface_f5cc59b44264101a8c5908002b2f8426_1_1

import "testing"

// TestSyntaxID pins the abstract syntax: f5cc59b4-4264-101a-8c59-08002b2f8426 v1.1.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	u := s.UUID
	if u.A != 0xf5cc59b4 || u.B != 0x4264 || u.C != 0x101a || u.D != 0x8c59 || u.E != 0x08002b2f8426 {
		t.Errorf("UUID = %s, want f5cc59b4-4264-101a-8c59-08002b2f8426", u.ToFormatD())
	}
	if s.MajorVersion != 1 || s.MinorVersion != 1 {
		t.Errorf("version = %d.%d, want 1.1", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport: FRS has no named pipe (ncacn_ip_tcp only).
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (FRS is ncacn_ip_tcp only)", PipeName)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent, that the
// four on-the-wire opnums (0-3) are present exactly once, and that the "not used on the
// wire" opnums (4-10) are absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	wireOpnums := []uint16{0, 1, 2, 3}
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
	for _, gap := range []uint16{4, 5, 6, 7, 8, 9, 10} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("opnum %d (%q) is NotUsedOnWire and must be absent", gap, name)
		}
	}
}

// TestStatusString spot-checks the mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:           "ERROR_SUCCESS",
		ErrorAccessDenied:       "ERROR_ACCESS_DENIED",
		ErrorCallNotImplemented: "ERROR_CALL_NOT_IMPLEMENTED",
		0xdeadbeef:              "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
}
