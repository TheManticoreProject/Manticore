package rpcinterface_76f03f96cdfd44fca22c64950a001209_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestStatusCodesResolveThroughWin32 pins that the success code [MS-PAR] 3.1.4 documents
// for IRemoteWinspool resolves through the shared [MS-ERREF] 2.2 table under the name the
// specification gives it, that the print-specific codes the interface never enumerated
// now render by name rather than as undecoded hex, and that a value the specification
// does not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	// The only code the interface used to declare. [MS-ERREF] 2.2 lists zero twice, as
	// ERROR_SUCCESS and as NERR_Success; the methods return a Win32 error rather than a
	// NET_API_STATUS, so ERROR_SUCCESS is the name used here and the one String renders.
	if got := win32.WIN32_ERROR(0x00000000).String(); got != "ERROR_SUCCESS" {
		t.Errorf("win32.WIN32_ERROR(0x00000000).String() = %q, want ERROR_SUCCESS", got)
	}
	if code, defined := win32.FromName("ERROR_SUCCESS"); !defined || uint32(code) != 0x00000000 {
		t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x00000000, true", "ERROR_SUCCESS", uint32(code), defined)
	}
	if win32.NERR_Success != win32.ERROR_SUCCESS {
		t.Errorf("NERR_Success = 0x%08x, want the same value as ERROR_SUCCESS", uint32(win32.NERR_Success))
	}

	// Codes outside the one-value subset: each of these used to print as bare hex in a
	// "RpcAsync... failed" message.
	beyondSubset := map[uint32]string{
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x00000709: "ERROR_INVALID_PRINTER_NAME",
		0x0000070A: "ERROR_PRINTER_ALREADY_EXISTS",
		0x00000BB9: "ERROR_PRINTER_DRIVER_IN_USE",
		0x00000BBB: "ERROR_SPL_NO_STARTDOC",
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
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0x76f03f96 || id.UUID.B != 0xcdfd || id.UUID.C != 0x44fc ||
		id.UUID.D != 0xa22c || id.UUID.E != 0x64950a001209 {
		t.Errorf("SyntaxID UUID = %+v, want 76f03f96-cdfd-44fc-a22c-64950a001209", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}
