package rpcinterface_99fcfec45260101bbbcb00aa0021347a_0_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of IObjectExporter.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "99fcfec4-5260-101b-bbcb-00aa0021347a" {
		t.Fatalf("UUID = %s, want 99fcfec4-5260-101b-bbcb-00aa0021347a", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the six opnums and that OpnumToName/NameToOpnum stay in sync.
func TestOpnums(t *testing.T) {
	want := map[uint16]string{
		0: "ResolveOxid",
		1: "SimplePing",
		2: "ComplexPing",
		3: "ServerAlive",
		4: "ResolveOxid2",
		5: "ServerAlive2",
	}
	if len(OpnumToName) != len(want) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(want))
	}
	for op, name := range want {
		if OpnumToName[op] != name {
			t.Fatalf("OpnumToName[%d] = %q, want %q", op, OpnumToName[op], name)
		}
		if NameToOpnum[name] != op {
			t.Fatalf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("map sizes differ: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering of the documented codes and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess: "RPC_S_OK",
		OrInvalidOxid: "OR_INVALID_OXID",
		OrInvalidOid:  "OR_INVALID_OID",
		OrInvalidSet:  "OR_INVALID_SET",
		0xdeadbeef:    "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Fatalf("StatusString(0x%08x) = %s, want %s", code, got, want)
		}
	}
}
