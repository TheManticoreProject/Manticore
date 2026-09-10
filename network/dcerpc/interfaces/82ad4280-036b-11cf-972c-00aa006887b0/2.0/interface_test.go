package rpcinterface_82ad4280036b11cf972c00aa006887b0_2_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the inetinfo interface
// (82ad4280-036b-11cf-972c-00aa006887b0 v2.0, [MS-IRP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "82ad4280-036b-11cf-972c-00aa006887b0" {
		t.Errorf("UUID = %s, want 82ad4280-036b-11cf-972c-00aa006887b0", got)
	}
	if s.MajorVersion != 2 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 2.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that all 16 on-the-wire opnums (0..15) are covered. Opnums 16 and 17
// (Opnum16NotUsedOnWire / Opnum17NotUsedOnWire) are intentionally absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 16 {
		t.Fatalf("OpnumToName has %d entries, want 16 (opnums 0..15)", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for op := uint16(0); op < 16; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	if _, ok := OpnumToName[16]; ok {
		t.Errorf("opnum 16 is not used on the wire and must be absent")
	}
}

// TestStatusCodesResolveThroughWin32 pins that the one value this descriptor used to
// declare resolves through the shared [MS-ERREF] 2.2 table under its specification name,
// that the generic Win32 failures [MS-IRP] section 3.1.4 leaves unenumerated now render
// by name rather than as undecoded hex, and that a value the specification does not
// define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	// The whole of the old subset: success. [MS-ERREF] 2.2 carries two names on this
	// row, ERROR_SUCCESS and the NERR_Success alias; String renders the canonical
	// ERROR_SUCCESS, which is the name the descriptor used.
	if got := win32.WIN32_ERROR(0x00000000).String(); got != "ERROR_SUCCESS" {
		t.Errorf("win32.WIN32_ERROR(0x00000000).String() = %q, want ERROR_SUCCESS", got)
	}
	if resolved, defined := win32.FromName("ERROR_SUCCESS"); !defined || uint32(resolved) != 0x00000000 {
		t.Errorf("win32.FromName(ERROR_SUCCESS) = 0x%08x, %v; want 0x00000000, true", uint32(resolved), defined)
	}

	// [MS-IRP] enumerates no per-method code set, so every failure this interface can
	// report was outside the one-value subset and used to come out as bare hex. These
	// resolve by name now.
	outsideOldSubset := map[uint32]string{
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x0000001F: "ERROR_GEN_FAILURE",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x000000EA: "ERROR_MORE_DATA",
	}
	for code, name := range outsideOldSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestPipeName pins the [MS-IRP] 2.1.1 well-known endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\PIPE\inetinfo` {
		t.Errorf("PipeName = %q, want \\PIPE\\inetinfo", PipeName)
	}
}
