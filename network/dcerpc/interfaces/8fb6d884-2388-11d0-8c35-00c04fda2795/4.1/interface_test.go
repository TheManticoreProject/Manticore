package rpcinterface_8fb6d884238811d08c3500c04fda2795_4_1

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the W32Time interface
// (8fb6d884-2388-11d0-8c35-00c04fda2795 v4.1, [MS-W32T]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "8fb6d884-2388-11d0-8c35-00c04fda2795" {
		t.Errorf("UUID = %s, want 8fb6d884-2388-11d0-8c35-00c04fda2795", got)
	}
	if s.MajorVersion != 4 || s.MinorVersion != 1 {
		t.Errorf("version = %d.%d, want 4.1", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover every opnum 0..7.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 8 {
		t.Fatalf("OpnumToName has %d entries, want 8", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, n := range OpnumToName {
		if NameToOpnum[n] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, n, NameToOpnum[n])
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the ten Win32 codes [MS-W32T] section
// 3.2.4 documents for W32Time resolve through the shared [MS-ERREF] 2.2 table under
// their specification names, that codes the interface never enumerated now render by
// name rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x00000426: "ERROR_SERVICE_NOT_ACTIVE",
		0x000005B4: "ERROR_TIMEOUT",
	}
	if len(documented) != 10 {
		t.Fatalf("documented table has %d entries, want the 10 codes the descriptor declared", len(documented))
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// These sit outside the subset the descriptor used to declare, two of them a single
	// value away from ERROR_SERVICE_NOT_ACTIVE, so a method returning one of them used to
	// render as undecoded hex. They resolve by name now.
	outsideOldSubset := map[uint32]string{
		0x00000006: "ERROR_INVALID_HANDLE",
		0x0000001F: "ERROR_GEN_FAILURE",
		0x00000425: "ERROR_SERVICE_CANNOT_ACCEPT_CTRL",
		0x00000427: "ERROR_FAILED_SERVICE_CONTROLLER_CONNECT",
		0x0000045B: "ERROR_SHUTDOWN_IN_PROGRESS",
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

// TestPipeName pins the [MS-W32T] section 2.1 well-known endpoints.
func TestPipeName(t *testing.T) {
	if PipeName != `\W32TIME` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\W32TIME`)
	}
	if PipeNameAlt != `\W32TIME_ALT` {
		t.Errorf("PipeNameAlt = %q, want %q", PipeNameAlt, `\W32TIME_ALT`)
	}
}
