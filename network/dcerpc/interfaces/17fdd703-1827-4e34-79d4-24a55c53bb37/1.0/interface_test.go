package rpcinterface_17fdd70318274e3479d424a55c53bb37_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "17fdd703-1827-4e34-79d4-24a55c53bb37" {
		t.Fatalf("UUID = %s, want 17fdd703-1827-4e34-79d4-24a55c53bb37", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the four opnums and their name mappings round-trip.
func TestOpnums(t *testing.T) {
	want := map[uint16]string{
		0: "NetrMessageNameAdd",
		1: "NetrMessageNameEnum",
		2: "NetrMessageNameGetInfo",
		3: "NetrMessageNameDel",
	}
	for op, name := range want {
		if OpnumToName[op] != name || NameToOpnum[name] != op {
			t.Errorf("opnum %d <-> %q mapping is inconsistent", op, name)
		}
	}
	if len(OpnumToName) != len(NameToOpnum) || len(OpnumToName) != len(want) {
		t.Fatalf("map sizes differ: OpnumToName=%d NameToOpnum=%d want=%d", len(OpnumToName), len(NameToOpnum), len(want))
	}
}

// TestStatusCodesResolveThroughWin32 pins that the NET_API_STATUS codes [MS-MSRP]
// 3.1.4 documents for msgsvc resolve through the shared [MS-ERREF] 2.2 table under
// their specification names, that a code the interface never enumerated now renders
// by name rather than as undecoded hex, and that a value the specification does not
// define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007B: "ERROR_INVALID_NAME",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x0000084B: "NERR_BufTooSmall",
		0x00000858: "NERR_NetworkError",
		0x0000085C: "NERR_InternalError",
		0x000008E4: "NERR_AlreadyExists",
		0x000008E5: "NERR_TooManyNames",
		0x000008E6: "NERR_DelComputerName",
		0x000008EB: "NERR_NameInUse",
		0x000008ED: "NERR_NotLocalName",
		0x000008F9: "NERR_DuplicateName",
		0x000008FB: "NERR_IncompleteDel",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// Success carries two names in [MS-ERREF] 2.2, ERROR_SUCCESS and NERR_Success, and
	// the method stubs compare against win32.NERR_Success because [MS-MSRP] 3.1.4
	// documents NET_API_STATUS returns. Both names resolve to the same value, which
	// renders under the canonical ERROR_SUCCESS.
	if resolved, defined := win32.FromName("NERR_Success"); !defined || uint32(resolved) != 0x00000000 {
		t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x00000000, true", "NERR_Success", uint32(resolved), defined)
	}
	if win32.NERR_Success != win32.ERROR_SUCCESS {
		t.Errorf("win32.NERR_Success = 0x%08x, want the same value as win32.ERROR_SUCCESS", uint32(win32.NERR_Success))
	}

	// NERR_MsgNotStarted, the Messenger service not being started, was outside the
	// subset this interface used to declare, so it rendered as hex; it resolves by name
	// now.
	if got := win32.WIN32_ERROR(0x000008EC).String(); got != "NERR_MsgNotStarted" {
		t.Errorf("win32.WIN32_ERROR(0x000008ec).String() = %q, want NERR_MsgNotStarted", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}
