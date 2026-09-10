package rpcinterface_8f09f000b7ed11cebbd200001a181cad_0_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the dimsvc interface
// (8f09f000-b7ed-11ce-bbd2-00001a181cad v0.0, [MS-RRASM]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "8f09f000-b7ed-11ce-bbd2-00001a181cad" {
		t.Errorf("UUID = %s, want 8f09f000-b7ed-11ce-bbd2-00001a181cad", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that all 53 on-the-wire opnums (0..52) are covered.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 53 {
		t.Fatalf("OpnumToName has %d entries, want 53 (opnums 0..52)", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for op := uint16(0); op < 53; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-RRASM] 3.1.4
// documents for dimsvc resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a code the interface never enumerated now renders by name
// rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000001: "ERROR_INVALID_FUNCTION",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x000000B7: "ERROR_ALREADY_EXISTS",
		0x00000103: "ERROR_NO_MORE_ITEMS",
		0x00000424: "ERROR_SERVICE_DOES_NOT_EXIST",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ERROR_MORE_DATA, ERROR_INVALID_LEVEL and ERROR_SERVICE_NOT_ACTIVE all sit outside
	// the twelve-code subset this interface used to declare, so a server returning any of
	// them rendered undecoded hex; they resolve by name now.
	outsideOldSubset := map[uint32]string{
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x000000EA: "ERROR_MORE_DATA",
		0x00000426: "ERROR_SERVICE_NOT_ACTIVE",
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

// TestPipeName pins the [MS-RRASM] 2.1 well-known endpoint shared with rasrpc.
func TestPipeName(t *testing.T) {
	if PipeName != `\ROUTER` {
		t.Errorf("PipeName = %q, want \\ROUTER", PipeName)
	}
}
