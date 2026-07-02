package rpcinterface_4fc742e04a1011cf827300aa004ae673_3_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorAccessDenied); got != "ERROR_ACCESS_DENIED" {
		t.Errorf("StatusString(0x5) = %q, want ERROR_ACCESS_DENIED", got)
	}
	if got := StatusString(ErrorNotFound); got != "ERROR_NOT_FOUND" {
		t.Errorf("StatusString(0x490) = %q, want ERROR_NOT_FOUND", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q, want hex fallback", got)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// MS-DFSNM defines 26 opnums, of which 3 (7, 8, 9) are "not used on the wire".
	if len(OpnumToName) != 23 {
		t.Errorf("OpnumToName has %d entries, want 23 on-the-wire methods", len(OpnumToName))
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
	// The 7/8/9 gap must not have leaked a mapping.
	for _, gap := range []uint16{7, 8, 9} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("OpnumToName[%d] = %q, want absent (not used on the wire)", gap, name)
		}
	}
	// Spot-check both directions, including a post-gap opnum.
	if OpnumToName[OpnumNetrDfsManagerGetVersion] != "NetrDfsManagerGetVersion" {
		t.Errorf("OpnumToName[0] = %q, want NetrDfsManagerGetVersion", OpnumToName[OpnumNetrDfsManagerGetVersion])
	}
	if NameToOpnum["NetrDfsAddFtRoot"] != OpnumNetrDfsAddFtRoot || OpnumNetrDfsAddFtRoot != 10 {
		t.Errorf("NameToOpnum[NetrDfsAddFtRoot] = %d, want 10", NameToOpnum["NetrDfsAddFtRoot"])
	}
	if NameToOpnum["NetrDfsGetSupportedNamespaceVersion"] != 25 {
		t.Errorf("NameToOpnum[NetrDfsGetSupportedNamespaceVersion] = %d, want 25", NameToOpnum["NetrDfsGetSupportedNamespaceVersion"])
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// 4fc742e0-4a10-11cf-8273-00aa004ae673, version 3.0.
	if id.UUID.A != 0x4fc742e0 || id.UUID.B != 0x4a10 || id.UUID.C != 0x11cf ||
		id.UUID.D != 0x8273 || id.UUID.E != 0x00aa004ae673 {
		t.Errorf("SyntaxID UUID = %+v, want 4fc742e0-4a10-11cf-8273-00aa004ae673", id.UUID)
	}
	if id.MajorVersion != 3 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 3.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestPipeName(t *testing.T) {
	if PipeName != `\netdfs` {
		t.Errorf("PipeName = %q, want \\netdfs", PipeName)
	}
}
