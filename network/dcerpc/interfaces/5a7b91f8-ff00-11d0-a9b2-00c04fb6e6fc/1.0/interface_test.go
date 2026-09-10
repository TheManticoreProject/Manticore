package rpcinterface_5a7b91f8ff0011d0a9b200c04fb6e6fc_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc" {
		t.Fatalf("UUID = %s, want 5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the single opnum and its name mapping round-trips.
func TestOpnums(t *testing.T) {
	if OpnumNetrSendMessage != 0 {
		t.Fatalf("OpnumNetrSendMessage = %d, want 0", OpnumNetrSendMessage)
	}
	if OpnumToName[0] != "NetrSendMessage" || NameToOpnum["NetrSendMessage"] != 0 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("map sizes differ: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughWin32 pins that the status codes [MS-MSRP] 3.2.4.1
// documents for NetrSendMessage resolve through the shared [MS-ERREF] 2.2 table under
// their specification names, that a code the interface never enumerated now renders by
// name rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x00000858: "NERR_NetworkError",
		0x000008E1: "NERR_NameNotFound",
		0x000008E8: "NERR_GrpMsgProcessor",
		0x000008E9: "NERR_PausedRemote",
		0x000008EA: "NERR_BadReceive",
		0x000008EB: "NERR_NameInUse",
		0x000008ED: "NERR_NotLocalName",
		0x000008F1: "NERR_TruncatedBroadcast",
		0x000008F9: "NERR_DuplicateName",
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
	// the method stub compares against win32.NERR_Success because [MS-MSRP] 3.2.4.1
	// documents NetrSendMessage's return values as NERR_* codes. Both names resolve to
	// the same value, which renders under the canonical ERROR_SUCCESS.
	if resolved, defined := win32.FromName("NERR_Success"); !defined || uint32(resolved) != 0x00000000 {
		t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x00000000, true", "NERR_Success", uint32(resolved), defined)
	}
	if win32.NERR_Success != win32.ERROR_SUCCESS {
		t.Errorf("win32.NERR_Success = 0x%08x, want the same value as win32.ERROR_SUCCESS", uint32(win32.NERR_Success))
	}

	// NERR_RemoteFull, the message alias table on the remote station being full, was
	// outside the subset this interface used to declare, so it rendered as hex; it
	// resolves by name now.
	if got := win32.WIN32_ERROR(0x000008EF).String(); got != "NERR_RemoteFull" {
		t.Errorf("win32.WIN32_ERROR(0x000008ef).String() = %q, want NERR_RemoteFull", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}
