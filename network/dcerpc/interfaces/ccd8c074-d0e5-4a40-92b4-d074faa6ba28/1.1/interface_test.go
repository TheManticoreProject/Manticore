package rpcinterface_ccd8c074d0e54a4092b4d074faa6ba28_1_1

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

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

// TestStatusCodesResolveThroughWin32 pins that the seven Win32 codes [MS-SWN] section
// 3.1.4 documents for the Witnessr* methods resolve through the shared [MS-ERREF] 2.2
// table under their specification names, that codes the descriptor never enumerated now
// render by name rather than as undecoded hex, and that a value the specification does
// not define still renders as hex. None of the seven is private to [MS-SWN]: each was
// looked up by value in the shared table, and the name the table records is the name the
// descriptor's comment carried.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x00000490: "ERROR_NOT_FOUND",
		0x0000051A: "ERROR_REVISION_MISMATCH",
		0x000005AA: "ERROR_NO_SYSTEM_RESOURCES",
		0x0000139F: "ERROR_INVALID_STATE",
	}
	if len(documented) != 7 {
		t.Fatalf("documented table has %d entries, want the 7 codes the descriptor declared", len(documented))
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// These sit outside the subset the descriptor used to declare, so a method returning
	// one of them used to render as undecoded hex: ERROR_INVALID_HANDLE for a context
	// handle the server no longer knows, ERROR_NO_MATCH a single value past the
	// ERROR_NOT_FOUND the subset did list, ERROR_NONPAGED_SYSTEM_RESOURCES a single value
	// past ERROR_NO_SYSTEM_RESOURCES, ERROR_TIMEOUT, and ERROR_CLUSTER_SHUTTING_DOWN a
	// single value before ERROR_INVALID_STATE. They resolve by name now. Naming them here
	// says only that they decode: the method stubs still treat every status other than
	// ERROR_SUCCESS as a failure, exactly as they did before.
	outsideOldSubset := map[uint32]string{
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000491: "ERROR_NO_MATCH",
		0x000005AB: "ERROR_NONPAGED_SYSTEM_RESOURCES",
		0x000005B4: "ERROR_TIMEOUT",
		0x0000139E: "ERROR_CLUSTER_SHUTTING_DOWN",
	}
	for code, name := range outsideOldSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
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
