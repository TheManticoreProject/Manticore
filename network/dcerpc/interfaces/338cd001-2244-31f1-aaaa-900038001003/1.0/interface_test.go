package rpcinterface_338cd001224431f1aaaa900038001003_1_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorFileNotFound); got != "ERROR_FILE_NOT_FOUND" {
		t.Errorf("StatusString(2) = %q", got)
	}
	if got := StatusString(ErrorMoreData); got != "ERROR_MORE_DATA" {
		t.Errorf("StatusString(234) = %q", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q, want hex", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0x338cd001 || id.UUID.B != 0x2244 || id.UUID.C != 0x31f1 ||
		id.UUID.D != 0xaaaa || id.UUID.E != 0x900038001003 {
		t.Errorf("SyntaxID UUID = %+v, want 338cd001-2244-31f1-aaaa-900038001003", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// 36 opnum slots (0-35), including the 5 reserved NotImplemented slots kept for
	// wire-numbering completeness.
	if len(OpnumToName) != 36 {
		t.Errorf("OpnumToName has %d entries, want 36 opnum slots", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumBaseRegOpenKey] != "BaseRegOpenKey" || OpnumBaseRegOpenKey != 15 {
		t.Errorf("OpnumBaseRegOpenKey = %d (%q)", OpnumBaseRegOpenKey, OpnumToName[OpnumBaseRegOpenKey])
	}
	if OpnumOpenLocalMachine != 2 || OpnumBaseRegQueryValue != 17 || OpnumBaseRegDeleteKeyEx != 35 {
		t.Errorf("opnum numbering off: OpenLocalMachine=%d QueryValue=%d DeleteKeyEx=%d",
			OpnumOpenLocalMachine, OpnumBaseRegQueryValue, OpnumBaseRegDeleteKeyEx)
	}
}
