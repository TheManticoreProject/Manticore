package rpcinterface_e35142354b0611d1ab0400c04fc2dcd2_4_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the drsuapi interface
// (e3514235-4b06-11d1-ab04-00c04fc2dcd2 v4.0, [MS-DRSR]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "e3514235-4b06-11d1-ab04-00c04fc2dcd2" {
		t.Errorf("UUID = %s, want e3514235-4b06-11d1-ab04-00c04fc2dcd2", got)
	}
	if s.MajorVersion != 4 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 4.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses, so the
// two maps never drift, and that the on-the-wire opnum range (0..30) is fully covered.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 31 {
		t.Fatalf("OpnumToName has %d entries, want 31 (opnums 0..30)", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for op := uint16(0); op <= 30; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the status codes this interface
// documents render under their [MS-ERREF] 2.2 names through the shared win32
// table, that DRA codes the package's own curated subset never listed —
// ERROR_DS_DRA_SCHEMA_MISMATCH among them — now render by name rather than as a
// bare hex value, and that a value the specification does not define still falls
// back to hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[win32.WIN32_ERROR]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x000020F5: "ERROR_DS_DRA_INVALID_PARAMETER",
		0x000020F6: "ERROR_DS_DRA_BUSY",
		0x000020F7: "ERROR_DS_DRA_BAD_DN",
		0x000020F8: "ERROR_DS_DRA_BAD_NC",
		0x000020F9: "ERROR_DS_DRA_DN_EXISTS",
		0x000020FA: "ERROR_DS_DRA_INTERNAL_ERROR",
		0x000020FE: "ERROR_DS_DRA_OUT_OF_MEM",
		0x00002105: "ERROR_DS_DRA_ACCESS_DENIED",
	}
	for code, want := range documented {
		if got := code.String(); got != want {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", uint32(code), got, want)
		}
	}

	// DRA codes outside the curated subset this package used to carry. These came
	// back as a bare hex value before the migration and now resolve by name.
	beyondTheSubset := map[win32.WIN32_ERROR]string{
		0x000020E2: "ERROR_DS_DRA_SCHEMA_MISMATCH",
		0x00002108: "ERROR_DS_DRA_SOURCE_DISABLED",
		0x0000210C: "ERROR_DS_DRA_MISSING_PARENT",
		0x00002160: "ERROR_DS_DRA_EARLIER_SCHEMA_CONFLICT",
	}
	for code, want := range beyondTheSubset {
		if got := code.String(); got != want {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", uint32(code), got, want)
		}
	}

	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xDEADBEEF).String() = %q, want 0xdeadbeef", got)
	}
}
