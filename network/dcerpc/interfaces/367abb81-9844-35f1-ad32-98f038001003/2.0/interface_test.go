package rpcinterface_367abb81984435f1ad3298f038001003_2_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorServiceDoesNotExist); got != "ERROR_SERVICE_DOES_NOT_EXIST" {
		t.Errorf("StatusString(0x424) = %q", got)
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
	if id.UUID.A != 0x367abb81 || id.UUID.B != 0x9844 || id.UUID.C != 0x35f1 ||
		id.UUID.D != 0xad32 || id.UUID.E != 0x98f038001003 {
		t.Errorf("SyntaxID UUID = %+v, want 367abb81-9844-35f1-ad32-98f038001003", id.UUID)
	}
	if id.MajorVersion != 2 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 2.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 50 {
		t.Errorf("OpnumToName has %d entries, want 50 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	// Spot-check a few opnum numbers against [MS-SCMR] 3.1.4.
	if OpnumToName[OpnumROpenSCManagerW] != "ROpenSCManagerW" || OpnumROpenSCManagerW != 15 {
		t.Errorf("OpnumROpenSCManagerW = %d (%q)", OpnumROpenSCManagerW, OpnumToName[OpnumROpenSCManagerW])
	}
	if OpnumToName[OpnumRCloseServiceHandle] != "RCloseServiceHandle" || OpnumRCloseServiceHandle != 0 {
		t.Errorf("OpnumRCloseServiceHandle = %d", OpnumRCloseServiceHandle)
	}
	if OpnumROpenSCManager2 != 64 {
		t.Errorf("OpnumROpenSCManager2 = %d, want 64", OpnumROpenSCManager2)
	}
}

// TestNotUsedOnWireOmitted asserts the 15 "not used on the wire" opnums are absent.
func TestNotUsedOnWireOmitted(t *testing.T) {
	for _, op := range []uint16{10, 22, 34, 43, 46, 52, 53, 54, 55, 57, 58, 59, 61, 62, 63} {
		if name, ok := OpnumToName[op]; ok {
			t.Errorf("opnum %d should be NotUsedOnWire but maps to %q", op, name)
		}
	}
}
