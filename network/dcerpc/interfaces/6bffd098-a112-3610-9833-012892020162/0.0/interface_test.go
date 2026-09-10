package rpcinterface_6bffd098a11236109833012892020162_0_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "6bffd098-a112-3610-9833-012892020162" {
		t.Fatalf("UUID = %s, want 6bffd098-a112-3610-9833-012892020162", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the single on-the-wire opnum and the name mapping round-trip.
func TestOpnums(t *testing.T) {
	if OpnumI_BrowserrQueryOtherDomains != 2 {
		t.Fatalf("OpnumI_BrowserrQueryOtherDomains = %d, want 2", OpnumI_BrowserrQueryOtherDomains)
	}
	if OpnumToName[2] != "I_BrowserrQueryOtherDomains" || NameToOpnum["I_BrowserrQueryOtherDomains"] != 2 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName/NameToOpnum size mismatch: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughWin32 pins that the six NET_API_STATUS codes this
// interface used to declare itself resolve through the shared [MS-ERREF] 2.2 table under
// their specification names, that a code the interface never enumerated now renders by
// name rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x000000EA: "ERROR_MORE_DATA",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// NERR_Success is the NET_API_STATUS name [MS-ERREF] 2.2 gives the same zero it also
	// names ERROR_SUCCESS, so the success comparison the stub makes against
	// win32.NERR_Success is the one the specification documents.
	if uint32(win32.NERR_Success) != 0x00000000 {
		t.Errorf("win32.NERR_Success = 0x%08x, want 0x00000000", uint32(win32.NERR_Success))
	}
	if resolved, defined := win32.FromName("NERR_Success"); !defined || uint32(resolved) != 0x00000000 {
		t.Errorf("win32.FromName(\"NERR_Success\") = 0x%08x, %v; want 0x00000000, true", uint32(resolved), defined)
	}

	// ERROR_INVALID_HANDLE sits one value away from a code the subset did list and was
	// outside it, so it rendered as hex; it resolves by name now.
	if got := win32.WIN32_ERROR(0x00000006).String(); got != "ERROR_INVALID_HANDLE" {
		t.Errorf("win32.WIN32_ERROR(0x00000006).String() = %q, want ERROR_INVALID_HANDLE", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestPipeName pins the transport endpoint ([MS-BRWSA] 2.1).
func TestPipeName(t *testing.T) {
	if PipeName != `\browser` {
		t.Fatalf("PipeName = %q, want %q", PipeName, `\browser`)
	}
}
