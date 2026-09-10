package rpcinterface_f5cc59b44264101a8c5908002b2f8426_1_1

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax: f5cc59b4-4264-101a-8c59-08002b2f8426 v1.1.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	u := s.UUID
	if u.A != 0xf5cc59b4 || u.B != 0x4264 || u.C != 0x101a || u.D != 0x8c59 || u.E != 0x08002b2f8426 {
		t.Errorf("UUID = %s, want f5cc59b4-4264-101a-8c59-08002b2f8426", u.ToFormatD())
	}
	if s.MajorVersion != 1 || s.MinorVersion != 1 {
		t.Errorf("version = %d.%d, want 1.1", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport: FRS has no named pipe (ncacn_ip_tcp only).
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (FRS is ncacn_ip_tcp only)", PipeName)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent, that the
// four on-the-wire opnums (0-3) are present exactly once, and that the "not used on the
// wire" opnums (4-10) are absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	wireOpnums := []uint16{0, 1, 2, 3}
	if len(OpnumToName) != len(wireOpnums) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(wireOpnums))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	for _, op := range wireOpnums {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("on-the-wire opnum %d missing from OpnumToName", op)
		}
	}
	for _, gap := range []uint16{4, 5, 6, 7, 8, 9, 10} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("opnum %d (%q) is NotUsedOnWire and must be absent", gap, name)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the three status codes this descriptor
// used to declare resolve through the shared [MS-ERREF] 2.2 table under the names the
// specification gives them, that FRS codes the descriptor never enumerated now render by
// name rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	// The retired subset: StatusSuccess, ErrorAccessDenied, ErrorCallNotImplemented.
	retired := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000078: "ERROR_CALL_NOT_IMPLEMENTED",
	}
	for code, name := range retired {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// The NtFrs service errors [MS-ERREF] 2.2 carries at 0x00001F41..0x00001F51 sat
	// outside the subset this interface declared, so they rendered as bare hex; they
	// resolve by name now. These two are the promotion-path failures FrsRpc reports.
	beyondSubset := map[uint32]string{
		0x00001F49: "FRS_ERR_PARENT_INSUFFICIENT_PRIV",
		0x00001F4A: "FRS_ERR_PARENT_AUTHENTICATION",
	}
	for code, name := range beyondSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xdeadbeef).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want 0xdeadbeef", got)
	}
}
