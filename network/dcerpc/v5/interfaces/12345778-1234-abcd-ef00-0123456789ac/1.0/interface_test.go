package rpcinterface_123457781234abcdef000123456789ac_1_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusMoreEntries); got != "STATUS_MORE_ENTRIES" {
		t.Errorf("StatusString(0x105) = %q", got)
	}
	if got := StatusString(StatusSuccess); got != "STATUS_SUCCESS" {
		t.Errorf("StatusString(0) = %q", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q, want hex", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0x12345778 || id.UUID.B != 0x1234 || id.UUID.C != 0xabcd ||
		id.UUID.D != 0xef00 || id.UUID.E != 0x0123456789ac {
		t.Errorf("SyntaxID UUID = %+v, want 12345778-1234-abcd-ef00-0123456789ac", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 64 {
		t.Errorf("OpnumToName has %d entries, want 64 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumSamrConnect5] != "SamrConnect5" {
		t.Errorf("OpnumToName[64] = %q", OpnumToName[OpnumSamrConnect5])
	}
}
