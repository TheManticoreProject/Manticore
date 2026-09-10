package rpcinterface_d049b186814f11d19a3c00c04fc9b232_1_1

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax: d049b186-814f-11d1-9a3c-00c04fc9b232 v1.1.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	u := s.UUID
	if u.A != 0xd049b186 || u.B != 0x814f || u.C != 0x11d1 || u.D != 0x9a3c || u.E != 0x00c04fc9b232 {
		t.Errorf("UUID = %s, want d049b186-814f-11d1-9a3c-00c04fc9b232", u.ToFormatD())
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
// six on-the-wire opnums (4, 5, 7, 8, 9, 10) are present exactly once, and that the
// "not used on the wire" opnums (0, 1, 2, 3, 6) are absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	wireOpnums := []uint16{4, 5, 7, 8, 9, 10}
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
	for _, gap := range []uint16{0, 1, 2, 3, 6} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("opnum %d (%q) is NotUsedOnWire and must be absent", gap, name)
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the two status codes this descriptor
// used to declare resolve through the shared [MS-ERREF] 2.2 table under the names the
// specification gives them, that the NtFrs service errors the descriptor never
// enumerated now render by name rather than as undecoded hex, and that a value the
// specification does not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	// The retired subset: StatusSuccess and ErrorAccessDenied.
	retired := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
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
	// resolve by name now. These are the failures an FRSAPI call reports.
	serviceErrors := map[uint32]string{
		0x00001F41: "FRS_ERR_INVALID_API_SEQUENCE",
		0x00001F44: "FRS_ERR_INTERNAL_API",
		0x00001F46: "FRS_ERR_SERVICE_COMM",
		0x00001F47: "FRS_ERR_INSUFFICIENT_PRIV",
		0x00001F51: "FRS_ERR_INVALID_SERVICE_PARAMETER",
	}
	for code, name := range serviceErrors {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// The FRS_ERROR_* family [MS-FRS2] defines at 0x23xx is a different set of values
	// the shared table has no rows for, so nothing in this interface may be resolved
	// against it by name similarity.
	for _, code := range []uint32{0x00002342, 0x0000235A, 0x00002375} {
		if entry, defined := win32.Lookup(win32.WIN32_ERROR(code)); defined {
			t.Errorf("win32.Lookup(0x%08x) resolves to %q; [MS-ERREF] 2.2 has no row for it", code, entry.Name)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xdeadbeef).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want 0xdeadbeef", got)
	}
}
