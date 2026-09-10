package rpcinterface_5ca4a760ebb111cf861100a0245420ed_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the IcaApi interface
// (5ca4a760-ebb1-11cf-8611-00a0245420ed v1.0, [MS-TSTS]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "5ca4a760-ebb1-11cf-8611-00a0245420ed" {
		t.Errorf("UUID = %s, want 5ca4a760-ebb1-11cf-8611-00a0245420ed", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) == 0 {
		t.Fatal("OpnumToName is empty")
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

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-TSTS] 3.7 reports in
// the pResult [out] parameter resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a code this interface never enumerated now renders by name
// rather than as undecoded hex, and that a value the specification does not define still
// renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ERROR_CTX_WINSTATION_NOT_FOUND is the Terminal Services code a session lookup
	// fails with. It was outside the subset this interface used to declare, so it
	// rendered as hex; it resolves by name now.
	if got := win32.WIN32_ERROR(0x00001B6E).String(); got != "ERROR_CTX_WINSTATION_NOT_FOUND" {
		t.Errorf("win32.WIN32_ERROR(0x00001b6e).String() = %q, want ERROR_CTX_WINSTATION_NOT_FOUND", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}

// TestPipeName pins the [MS-TSTS] section 1.9 endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\Ctx_WinStation_API_service` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\Ctx_WinStation_API_service`)
	}
}
