package rpcinterface_0b1c217057324e0e8cd3d9b16f3b84d7_0_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the authzr interface
// (0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7 v0.0, [MS-RAA]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7" {
		t.Errorf("UUID = %s, want 0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// the anchor opnums resolve to the expected method names. authzr exposes 7 contiguous
// opnums (0..6); none are "not used on the wire".
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 7 {
		t.Fatalf("OpnumToName has %d entries, want 7", len(OpnumToName))
	}
	if OpnumToName[OpnumAuthzrFreeContext] != "AuthzrFreeContext" {
		t.Errorf("opnum 0 = %q, want AuthzrFreeContext", OpnumToName[OpnumAuthzrFreeContext])
	}
	if OpnumToName[OpnumAuthzrModifySids] != "AuthzrModifySids" {
		t.Errorf("opnum 6 = %q, want AuthzrModifySids", OpnumToName[OpnumAuthzrModifySids])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestOpnumContiguous checks the opnum space is exactly 0..6 with no gaps.
func TestOpnumContiguous(t *testing.T) {
	for op := uint16(0); op <= 6; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	if _, ok := OpnumToName[7]; ok {
		t.Errorf("opnum 7 should not exist")
	}
}

// TestStatusCodesResolveThroughWin32 pins that the eight Win32 codes [MS-RAA] section
// 3.1.4 documents for authzr resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that codes the descriptor never enumerated now render by name
// rather than as undecoded hex, and that a value the specification does not define still
// renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x00000548: "ERROR_INVALID_SERVER_STATE",
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

	// These sit outside the subset the descriptor used to declare, so a method returning
	// one of them used to render as undecoded hex: ERROR_NONE_MAPPED is what a SID that
	// resolves to no account produces in AuthzrInitializeContextFromSid, and
	// ERROR_INVALID_DOMAIN_STATE is a single value past ERROR_INVALID_SERVER_STATE, which
	// the subset did list. They resolve by name now.
	outsideOldSubset := map[uint32]string{
		0x0000000E: "ERROR_OUTOFMEMORY",
		0x00000522: "ERROR_PRIVILEGE_NOT_HELD",
		0x00000534: "ERROR_NONE_MAPPED",
		0x00000549: "ERROR_INVALID_DOMAIN_STATE",
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

// TestPipeName documents that this interface has no named pipe: [MS-RAA] 2.1 uses
// ncacn_ip_tcp with a dynamic endpoint, so PipeName is intentionally empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}
