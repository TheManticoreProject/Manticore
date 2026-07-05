package rpcinterface_ccd8c074d0e54a4092b4d074faa6ba28_1_1

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the Witness interface
// (ccd8c074-d0e5-4a40-92b4-d074faa6ba28 v1.1, [MS-SWN]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "ccd8c074-d0e5-4a40-92b4-d074faa6ba28" {
		t.Errorf("UUID = %s, want ccd8c074-d0e5-4a40-92b4-d074faa6ba28", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 1 {
		t.Errorf("version = %d.%d, want 1.1", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// the anchor opnums resolve to the expected method names. Witness exposes 6 contiguous
// opnums (0..5); none are "not used on the wire".
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 6 {
		t.Fatalf("OpnumToName has %d entries, want 6", len(OpnumToName))
	}
	if OpnumToName[OpnumWitnessrGetInterfaceList] != "WitnessrGetInterfaceList" {
		t.Errorf("opnum 0 = %q, want WitnessrGetInterfaceList", OpnumToName[OpnumWitnessrGetInterfaceList])
	}
	if OpnumToName[OpnumWitnessrUnRegisterEx] != "WitnessrUnRegisterEx" {
		t.Errorf("opnum 5 = %q, want WitnessrUnRegisterEx", OpnumToName[OpnumWitnessrUnRegisterEx])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestOpnumContiguous checks the opnum space is exactly 0..5 with no gaps.
func TestOpnumContiguous(t *testing.T) {
	for op := uint16(0); op <= 5; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	if _, ok := OpnumToName[6]; ok {
		t.Errorf("opnum 6 should not exist")
	}
}

// TestStatusString checks known Win32 mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:           "ERROR_SUCCESS",
		StatusAccessDenied:      "ERROR_ACCESS_DENIED",
		StatusInvalidParameter:  "ERROR_INVALID_PARAMETER",
		StatusNotFound:          "ERROR_NOT_FOUND",
		StatusRevisionMismatch:  "ERROR_REVISION_MISMATCH",
		StatusNoSystemResources: "ERROR_NO_SYSTEM_RESOURCES",
		StatusInvalidState:      "ERROR_INVALID_STATE",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestProtocolVersions pins the two documented Witness protocol versions ([MS-SWN]
// 2.2.2.3): a mismatched Version parameter makes the server return ERROR_REVISION_MISMATCH.
func TestProtocolVersions(t *testing.T) {
	if WitnessVersionV1 != 0x00010001 {
		t.Errorf("WitnessVersionV1 = 0x%08x, want 0x00010001", WitnessVersionV1)
	}
	if WitnessVersionV2 != 0x00020000 {
		t.Errorf("WitnessVersionV2 = 0x%08x, want 0x00020000", WitnessVersionV2)
	}
}

// TestPipeName documents that this interface has no named pipe: [MS-SWN] 2.1 uses
// ncacn_ip_tcp with a dynamic endpoint, so PipeName is intentionally empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}
