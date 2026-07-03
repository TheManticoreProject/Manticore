package rpcinterface_45f52c287f9f101ab52b08002b2efabe_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the winsif interface
// (45f52c28-7f9f-101a-b52b-08002b2efabe v1.0, [MS-RAIW]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "45f52c28-7f9f-101a-b52b-08002b2efabe" {
		t.Errorf("UUID = %s, want 45f52c28-7f9f-101a-b52b-08002b2efabe", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that the anchor opnums resolve to the expected method names. winsif exposes 22
// contiguous opnums (0..21); none are "not used on the wire".
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 22 {
		t.Fatalf("OpnumToName has %d entries, want 22", len(OpnumToName))
	}
	if OpnumToName[OpnumR_WinsRecordAction] != "R_WinsRecordAction" {
		t.Errorf("opnum 0 = %q, want R_WinsRecordAction", OpnumToName[OpnumR_WinsRecordAction])
	}
	if OpnumToName[OpnumR_WinsDoScavengingNew] != "R_WinsDoScavengingNew" {
		t.Errorf("opnum 21 = %q, want R_WinsDoScavengingNew", OpnumToName[OpnumR_WinsDoScavengingNew])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestOpnumContiguous checks the opnum space is exactly 0..21 with no gaps.
func TestOpnumContiguous(t *testing.T) {
	for op := uint16(0); op <= 21; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	if _, ok := OpnumToName[22]; ok {
		t.Errorf("opnum 22 should not exist")
	}
}

// TestStatusString checks known Win32 mnemonics, a WINS-specific code, and the hex
// fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:       "ERROR_SUCCESS",
		ErrorAccessDenied:   "ERROR_ACCESS_DENIED",
		ErrorWinsInternal:   "ERROR_WINS_INTERNAL",
		ErrorRplNotAllowed:  "ERROR_RPL_NOT_ALLOWED",
		ErrorRecNonExistent: "ERROR_REC_NON_EXISTENT",
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

// TestPipeName pins the shared WINS named pipe ([MS-RAIW] 2.1, Standards Assignments).
func TestPipeName(t *testing.T) {
	if PipeName != `\WinsPipe` {
		t.Errorf("PipeName = %q, want \\WinsPipe", PipeName)
	}
}
