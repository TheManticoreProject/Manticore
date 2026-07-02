package rpcinterface_906b0ce0c70b1067b31700dd010662da_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "906b0ce0-c70b-1067-b317-00dd010662da" {
		t.Fatalf("UUID = %s, want 906b0ce0-c70b-1067-b317-00dd010662da", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the eight opnums are contiguous 0..7 and the name maps round-trip.
func TestOpnums(t *testing.T) {
	want := map[uint16]string{
		0: "Poke",
		1: "BuildContext",
		2: "NegotiateResources",
		3: "SendReceive",
		4: "TearDownContext",
		5: "BeginTearDown",
		6: "PokeW",
		7: "BuildContextW",
	}
	if len(OpnumToName) != len(want) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(want))
	}
	for op, name := range want {
		if OpnumToName[op] != name {
			t.Errorf("OpnumToName[%d] = %q, want %q", op, OpnumToName[op], name)
		}
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("map sizes differ: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering of the recognized HRESULTs and the hex
// fallback for everything else.
func TestStatusString(t *testing.T) {
	cases := []struct {
		code uint32
		want string
	}{
		{StatusSuccess, "S_OK"},
		{StatusFail, "E_FAIL"},
		{StatusOutOfMemory, "E_OUTOFMEMORY"},
		{StatusInvalidArg, "E_INVALIDARG"},
		{0xDEADBEEF, "0xdeadbeef"},
	}
	for _, c := range cases {
		if got := StatusString(c.code); got != c.want {
			t.Errorf("StatusString(0x%08x) = %s, want %s", c.code, got, c.want)
		}
	}
}
