package rpcinterface_e1af83085d1f11c991a408002b14a0fa_3_0

import "testing"

func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if s.MajorVersion != 3 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 3.0", s.MajorVersion, s.MinorVersion)
	}
	want := "e1af8308-5d1f-11c9-91a4-08002b14a0fa"
	if got := s.UUID.ToFormatD(); got != want {
		t.Errorf("UUID = %s, want %s", got, want)
	}
}

func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:          "rpc_s_ok",
		EptStatusNotRegistered: "ept_s_not_registered",
		0xdeadbeef:             "0xdeadbeef",
	}
	for status, want := range cases {
		if got := StatusString(status); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", status, got, want)
		}
	}
}

func TestOpnumNameRoundTrip(t *testing.T) {
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	if OpnumEptMap != 3 {
		t.Errorf("OpnumEptMap = %d, want 3", OpnumEptMap)
	}
}
