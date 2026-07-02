package rpcinterface_b9785960524f11df8b6d83dcded72085_1_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "S_OK" {
		t.Errorf("StatusString(0) = %q, want S_OK", got)
	}
	if got := StatusString(0x80070005); got != "0x80070005" {
		t.Errorf("StatusString(unknown) = %q, want hex fallback", got)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// MS-GKDI defines a single on-the-wire method, GetKey (opnum 0).
	if len(OpnumToName) != 1 {
		t.Errorf("OpnumToName has %d entries, want 1 on-the-wire method", len(OpnumToName))
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
	if OpnumToName[OpnumGetKey] != "GetKey" {
		t.Errorf("OpnumToName[0] = %q, want GetKey", OpnumToName[OpnumGetKey])
	}
	if NameToOpnum["GetKey"] != OpnumGetKey || OpnumGetKey != 0 {
		t.Errorf("NameToOpnum[GetKey] = %d, want 0", NameToOpnum["GetKey"])
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// b9785960-524f-11df-8b6d-83dcded72085, version 1.0.
	if id.UUID.A != 0xb9785960 || id.UUID.B != 0x524f || id.UUID.C != 0x11df ||
		id.UUID.D != 0x8b6d || id.UUID.E != 0x83dcded72085 {
		t.Errorf("SyntaxID UUID = %+v, want b9785960-524f-11df-8b6d-83dcded72085", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestPipeName(t *testing.T) {
	if PipeName != `\lsass` {
		t.Errorf("PipeName = %q, want \\lsass", PipeName)
	}
}
