package rpcinterface_7c44d7d431d5424cbd5e2b3e1f323d22_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the dsaop interface
// (7c44d7d4-31d5-424c-bd5e-2b3e1f323d22 v1.0, [MS-DRSR]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "7c44d7d4-31d5-424c-bd5e-2b3e1f323d22" {
		t.Errorf("UUID = %s, want 7c44d7d4-31d5-424c-bd5e-2b3e1f323d22", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// both on-the-wire opnums (0..1) are covered.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 2 {
		t.Fatalf("OpnumToName has %d entries, want 2 (opnums 0..1)", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes dsaop methods return as
// their DWORD return value resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a code the interface never enumerated now renders by name
// rather than as undecoded hex, and that a value the specification does not define still
// renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
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

	// ERROR_INVALID_DOMAIN_ROLE and ERROR_DS_INTERNAL_FAILURE were outside the subset this
	// interface used to declare, so a demotion script failing on either rendered as hex;
	// they resolve by name now.
	if got := win32.WIN32_ERROR(0x0000054A).String(); got != "ERROR_INVALID_DOMAIN_ROLE" {
		t.Errorf("win32.WIN32_ERROR(0x0000054a).String() = %q, want ERROR_INVALID_DOMAIN_ROLE", got)
	}
	if got := win32.WIN32_ERROR(0x000020EE).String(); got != "ERROR_DS_INTERNAL_FAILURE" {
		t.Errorf("win32.WIN32_ERROR(0x000020ee).String() = %q, want ERROR_DS_INTERNAL_FAILURE", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}
