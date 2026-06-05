package rpcinterface_4b324fc8167001d312785a47bf6ee188_3_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(ERROR_MORE_DATA); got != "ERROR_MORE_DATA" {
		t.Errorf("StatusString(234) = %q", got)
	}
	if got := StatusString(NERR_Success); got != "NERR_Success" {
		t.Errorf("StatusString(0) = %q", got)
	}
	if got := StatusString(0x12345678); got != "305419896" {
		t.Errorf("StatusString(unknown) = %q, want decimal", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0x4b324fc8 || id.UUID.B != 0x1670 || id.UUID.C != 0x01d3 ||
		id.UUID.D != 0x1278 || id.UUID.E != 0x5a47bf6ee188 {
		t.Errorf("SyntaxID UUID = %+v, want 4b324fc8-1670-01d3-1278-5a47bf6ee188", id.UUID)
	}
	if id.MajorVersion != 3 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 3.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 47 {
		t.Errorf("OpnumToName has %d entries, want 47 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumNetrShareEnum] != "NetrShareEnum" {
		t.Errorf("OpnumToName[15] = %q", OpnumToName[OpnumNetrShareEnum])
	}
}
