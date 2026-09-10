package rpcinterface_3919286ab10c11d09ba800c04fd92ef5_0_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the dssetup interface
// (3919286a-b10c-11d0-9ba8-00c04fd92ef5 v0.0, [MS-DSSP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "3919286a-b10c-11d0-9ba8-00c04fd92ef5" {
		t.Errorf("UUID = %s, want 3919286a-b10c-11d0-9ba8-00c04fd92ef5", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport endpoint. dssetup shares \lsarpc ([MS-DSSP] 2.1).
func TestPipeName(t *testing.T) {
	if PipeName != `\lsarpc` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\lsarpc`)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// the single on-the-wire opnum (0) is covered; opnums 1..11 are NotUsedOnWire and absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 1 {
		t.Fatalf("OpnumToName has %d entries, want 1 (only opnum 0 is on the wire)", len(OpnumToName))
	}
	if OpnumToName[OpnumDsRolerGetPrimaryDomainInformation] != "DsRolerGetPrimaryDomainInformation" {
		t.Errorf("opnum 0 = %q, want DsRolerGetPrimaryDomainInformation", OpnumToName[OpnumDsRolerGetPrimaryDomainInformation])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the eight Win32 codes the descriptor
// declared for DsRolerGetPrimaryDomainInformation ([MS-DSSP] 3.2.5.1) resolve through the
// shared [MS-ERREF] 2.2 table under their specification names, that codes the interface
// never enumerated now render by name rather than as undecoded hex, and that a value the
// specification does not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x0000000E: "ERROR_OUTOFMEMORY",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x0000054A: "ERROR_INVALID_DOMAIN_ROLE",
		0x0000054B: "ERROR_NO_SUCH_DOMAIN",
	}
	if len(documented) != 8 {
		t.Fatalf("documented table has %d entries, want the 8 codes the descriptor declared", len(documented))
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
	// value either side of ERROR_INVALID_DOMAIN_ROLE and ERROR_NO_SUCH_DOMAIN, so a call
	// returning one of them used to render as undecoded hex. They resolve by name now.
	outsideOldSubset := map[uint32]string{
		0x00000006: "ERROR_INVALID_HANDLE",
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000549: "ERROR_INVALID_DOMAIN_STATE",
		0x0000054C: "ERROR_DOMAIN_EXISTS",
		0x0000054F: "ERROR_INTERNAL_ERROR",
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
