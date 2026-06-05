package rpcinterface_c681d488d85011d08c5200c04fd90f7e_1_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q", got)
	}
	if got := StatusString(ErrorAccessDenied); got != "ERROR_ACCESS_DENIED" {
		t.Errorf("StatusString(5) = %q", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q, want hex", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0xc681d488 || id.UUID.B != 0xd850 || id.UUID.C != 0x11d0 ||
		id.UUID.D != 0x8c52 || id.UUID.E != 0x00c04fd90f7e {
		t.Errorf("SyntaxID UUID = %+v, want c681d488-d850-11d0-8c52-00c04fd90f7e", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 20 {
		t.Errorf("OpnumToName has %d entries, want 20 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumEfsRpcOpenFileRaw] != "EfsRpcOpenFileRaw" {
		t.Errorf("OpnumToName[0] = %q", OpnumToName[OpnumEfsRpcOpenFileRaw])
	}
}
