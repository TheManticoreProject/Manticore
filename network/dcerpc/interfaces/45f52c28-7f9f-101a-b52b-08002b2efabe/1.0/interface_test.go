package rpcinterface_45f52c287f9f101ab52b08002b2efabe_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

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

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes winsif methods return
// ([MS-RAIW] 3.1.4) resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, including the WINS-specific block, that a code the interface never
// enumerated now renders by name rather than as undecoded hex, and that a value the
// specification does not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x00000FA0: "ERROR_WINS_INTERNAL",
		0x00000FA1: "ERROR_CAN_NOT_DEL_LOCAL_WINS",
		0x00000FA2: "ERROR_STATIC_INIT",
		0x00000FA3: "ERROR_INC_BACKUP",
		0x00000FA4: "ERROR_FULL_BACKUP",
		0x00000FA5: "ERROR_REC_NON_EXISTENT",
		0x00000FA6: "ERROR_RPL_NOT_ALLOWED",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// The WINS block 0x00000FA0-0x00000FA6 is the one range where a private subset could
	// have hidden a collision: had [MS-ERREF] 2.2 assigned these values to unrelated
	// names, routing them through the shared table would have renamed WINS failures. Every
	// row in the range carries its WINS name and the range ends at 0x00000FA6, so nothing
	// here is interface-specific. This assertion is what stops a later table change from
	// silently reassigning them.
	for code, name := range map[uint32]string{
		0x00000FA0: "ERROR_WINS_INTERNAL",
		0x00000FA1: "ERROR_CAN_NOT_DEL_LOCAL_WINS",
		0x00000FA2: "ERROR_STATIC_INIT",
		0x00000FA3: "ERROR_INC_BACKUP",
		0x00000FA4: "ERROR_FULL_BACKUP",
		0x00000FA5: "ERROR_REC_NON_EXISTENT",
		0x00000FA6: "ERROR_RPL_NOT_ALLOWED",
	} {
		entry, defined := win32.Lookup(win32.WIN32_ERROR(code))
		if !defined {
			t.Errorf("win32.Lookup(0x%08x) undefined, want %s", code, name)
			continue
		}
		if entry.Name != name {
			t.Errorf("win32.Lookup(0x%08x).Name = %q, want %q", code, entry.Name, name)
		}
	}
	if _, defined := win32.Lookup(0x00000FA7); defined {
		t.Errorf("win32.Lookup(0x00000fa7) is defined; the WINS block ends at 0x00000fa6")
	}

	// ERROR_MORE_DATA and ERROR_INVALID_HANDLE were outside the subset this interface used
	// to declare, so they rendered as hex; they resolve by name now.
	if got := win32.WIN32_ERROR(0x000000EA).String(); got != "ERROR_MORE_DATA" {
		t.Errorf("win32.WIN32_ERROR(0x000000ea).String() = %q, want ERROR_MORE_DATA", got)
	}
	if got := win32.WIN32_ERROR(0x00000006).String(); got != "ERROR_INVALID_HANDLE" {
		t.Errorf("win32.WIN32_ERROR(0x00000006).String() = %q, want ERROR_INVALID_HANDLE", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestPipeName pins the shared WINS named pipe ([MS-RAIW] 2.1, Standards Assignments).
func TestPipeName(t *testing.T) {
	if PipeName != `\WinsPipe` {
		t.Errorf("PipeName = %q, want \\WinsPipe", PipeName)
	}
}
