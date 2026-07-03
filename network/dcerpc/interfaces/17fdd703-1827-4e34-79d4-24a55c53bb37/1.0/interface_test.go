package rpcinterface_17fdd70318274e3479d424a55c53bb37_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "17fdd703-1827-4e34-79d4-24a55c53bb37" {
		t.Fatalf("UUID = %s, want 17fdd703-1827-4e34-79d4-24a55c53bb37", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the four opnums and their name mappings round-trip.
func TestOpnums(t *testing.T) {
	want := map[uint16]string{
		0: "NetrMessageNameAdd",
		1: "NetrMessageNameEnum",
		2: "NetrMessageNameGetInfo",
		3: "NetrMessageNameDel",
	}
	for op, name := range want {
		if OpnumToName[op] != name || NameToOpnum[name] != op {
			t.Errorf("opnum %d <-> %q mapping is inconsistent", op, name)
		}
	}
	if len(OpnumToName) != len(NameToOpnum) || len(OpnumToName) != len(want) {
		t.Fatalf("map sizes differ: OpnumToName=%d NameToOpnum=%d want=%d", len(OpnumToName), len(NameToOpnum), len(want))
	}
}

// TestStatusString verifies mnemonic rendering of documented codes and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:     "NERR_Success",
		ErrorAccessDenied: "ERROR_ACCESS_DENIED",
		ErrorInvalidLevel: "ERROR_INVALID_LEVEL",
		NerrBufTooSmall:   "NERR_BufTooSmall",
		NerrIncompleteDel: "NERR_IncompleteDel",
		0xdeadbeef:        "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
}
