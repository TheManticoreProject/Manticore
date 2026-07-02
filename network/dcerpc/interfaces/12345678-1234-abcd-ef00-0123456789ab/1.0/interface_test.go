package rpcinterface_123456781234abcdef000123456789ab_1_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorInvalidHandle); got != "ERROR_INVALID_HANDLE" {
		t.Errorf("StatusString(0x6) = %q, want ERROR_INVALID_HANDLE", got)
	}
	if got := StatusString(ErrorSplNoStartdoc); got != "ERROR_SPL_NO_STARTDOC" {
		t.Errorf("StatusString(0xE15) = %q, want ERROR_SPL_NO_STARTDOC", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q, want hex fallback", got)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// MS-RPRN defines 124 opnums, of which 36 are "not used on the wire" and omitted.
	if len(OpnumToName) != 88 {
		t.Errorf("OpnumToName has %d entries, want 88 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d (a duplicate name collapsed an entry)",
			len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	// Spot-check both directions, including a post-gap opnum.
	if OpnumToName[OpnumRpcOpenPrinter] != "RpcOpenPrinter" {
		t.Errorf("OpnumToName[1] = %q, want RpcOpenPrinter", OpnumToName[OpnumRpcOpenPrinter])
	}
	if NameToOpnum["RpcXcvData"] != OpnumRpcXcvData || OpnumRpcXcvData != 88 {
		t.Errorf("NameToOpnum[RpcXcvData] = %d, want 88", NameToOpnum["RpcXcvData"])
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// 12345678-1234-abcd-ef00-0123456789ab, version 1.0.
	if id.UUID.A != 0x12345678 || id.UUID.B != 0x1234 || id.UUID.C != 0xabcd ||
		id.UUID.D != 0xef00 || id.UUID.E != 0x0123456789ab {
		t.Errorf("SyntaxID UUID = %+v, want 12345678-1234-abcd-ef00-0123456789ab", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}
