package rpcinterface_2f5f6520ca461067b31900dd010662da_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID checks the abstract syntax matches [MS-TRP] Appendix A.2 (tapsrv):
// 2f5f6520-ca46-1067-b319-00dd010662da, version 1.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "2f5f6520-ca46-1067-b319-00dd010662da" {
		t.Errorf("UUID = %s, want 2f5f6520-ca46-1067-b319-00dd010662da", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent inverses and
// cover the three on-the-wire opnums.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 3 {
		t.Fatalf("OpnumToName has %d entries, want 3", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if OpnumClientAttach != 0 || OpnumClientRequest != 1 || OpnumClientDetach != 2 {
		t.Errorf("opnums = %d/%d/%d, want 0/1/2", OpnumClientAttach, OpnumClientRequest, OpnumClientDetach)
	}
}

// TestStatusCodesResolveThroughWin32 pins that the one code this descriptor used to
// declare for ClientAttach's return resolves through the shared [MS-ERREF] 2.2 table under
// its specification name, that Win32 failures the one-entry subset could not name now
// render by name rather than as undecoded hex, that the two values [MS-TRP] mentions for
// this return which lie outside [MS-ERREF] 2.2 still render as hex, and that a value the
// specification does not define does too.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
	}
	if len(documented) != 1 {
		t.Fatalf("documented table has %d entries, want the 1 code the descriptor declared", len(documented))
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ClientAttach's failures are "as specified in [MS-ERREF]", and success was the only
	// value the descriptor could name, so every one of them used to render as undecoded
	// hex. These resolve by name now.
	outsideOldSubset := map[uint32]string{
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x00000426: "ERROR_SERVICE_NOT_ACTIVE",
		0x000006BA: "RPC_S_SERVER_UNAVAILABLE",
	}
	for code, name := range outsideOldSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
	}

	// The two values [MS-TRP] 3.1.4.1 names for this return are not Win32 error codes:
	// LINEERR_OPERATIONFAILED is a TAPI code in the 0x8000xxxx block and -19 is a raw
	// negative. [MS-ERREF] 2.2 covers neither, so both stay hexadecimal exactly as they
	// did before and neither can be misnamed out of the shared table.
	outsideTheTable := map[uint32]string{
		0x80000048: "0x80000048", // LINEERR_OPERATIONFAILED
		0xFFFFFFED: "0xffffffed", // -19, client without administrator access
	}
	for code, hex := range outsideTheTable {
		if _, defined := win32.Lookup(win32.WIN32_ERROR(code)); defined {
			t.Errorf("win32 names 0x%08x; it is a TAPI value and the Win32 table must not claim it", code)
		}
		if got := win32.WIN32_ERROR(code).String(); got != hex {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, hex)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x80000005).String(); got != "0x80000005" {
		t.Errorf("win32.WIN32_ERROR(0x80000005).String() = %q, want hex", got)
	}
}

// TestPipeName pins the well-known tapsrv endpoint ([MS-TRP] 2.1).
func TestPipeName(t *testing.T) {
	if PipeName != `\tapsrv` {
		t.Errorf("PipeName = %q, want \\tapsrv", PipeName)
	}
}
