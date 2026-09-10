package rpcinterface_123456781234abcdef000123456789ab_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-RPRN] 2.2.5 documents
// for the print protocol resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that codes the interface never enumerated now render by name
// rather than as undecoded hex, and that a value the specification does not define still
// renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x0000007B: "ERROR_INVALID_NAME",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x000000EA: "ERROR_MORE_DATA",
		0x00000103: "ERROR_NO_MORE_ITEMS",
		0x00000704: "ERROR_UNKNOWN_PORT",
		0x00000705: "ERROR_UNKNOWN_PRINTER_DRIVER",
		0x00000706: "ERROR_UNKNOWN_PRINTPROCESSOR",
		0x00000709: "ERROR_INVALID_PRINTER_NAME",
		0x0000070A: "ERROR_PRINTER_ALREADY_EXISTS",
		0x0000070C: "ERROR_INVALID_DATATYPE",
		0x00000771: "ERROR_PRINTER_DELETED",
		0x00000772: "ERROR_INVALID_PRINTER_STATE",
		0x00000BB9: "ERROR_PRINTER_DRIVER_IN_USE",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// These print-protocol codes sat outside the subset this interface used to declare, so
	// they reached the caller as bare hex; the shared table names them.
	outsideOldSubset := map[uint32]string{
		0x00000703: "ERROR_PRINTER_DRIVER_ALREADY_INSTALLED",
		0x00000708: "ERROR_INVALID_PRIORITY",
		0x00000BB8: "ERROR_UNKNOWN_PRINT_MONITOR",
		0x00000BBA: "ERROR_SPOOL_FILE_NOT_FOUND",
		0x00000BBB: "ERROR_SPL_NO_STARTDOC",
		0x00000BBC: "ERROR_SPL_NO_ADDJOB",
	}
	for code, name := range outsideOldSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}

// TestStatusStringKeepsTheLocalCode pins the one value this interface still decodes
// itself. [MS-ERREF] 2.2 has no row for 0x00000E15 - it places ERROR_SPL_NO_STARTDOC at
// 0x00000BBB - so the shared table renders it as hex and StatusString names it.
func TestStatusStringKeepsTheLocalCode(t *testing.T) {
	if ErrorSplNoStartdoc != 0x00000E15 {
		t.Errorf("ErrorSplNoStartdoc = 0x%08x, want 0x00000e15", ErrorSplNoStartdoc)
	}
	if got := StatusString(ErrorSplNoStartdoc); got != "ERROR_SPL_NO_STARTDOC" {
		t.Errorf("StatusString(0x00000e15) = %q, want ERROR_SPL_NO_STARTDOC", got)
	}
	if got := win32.WIN32_ERROR(ErrorSplNoStartdoc).String(); got != "0x00000e15" {
		t.Errorf("win32.WIN32_ERROR(0x00000e15).String() = %q, want hex", got)
	}
	// Every other code defers to the shared table.
	if got := StatusString(0x00000BBB); got != "ERROR_SPL_NO_STARTDOC" {
		t.Errorf("StatusString(0x00000bbb) = %q, want ERROR_SPL_NO_STARTDOC", got)
	}
	if got := StatusString(0x00000BB8); got != "ERROR_UNKNOWN_PRINT_MONITOR" {
		t.Errorf("StatusString(0x00000bb8) = %q, want ERROR_UNKNOWN_PRINT_MONITOR", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(0x12345678) = %q, want hex fallback", got)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// MS-RPRN defines 124 opnums, of which 36 are "not used on the wire" and omitted.
	if len(OpnumToName) != 88 {
		t.Errorf("OpnumToName has %d entries, want 88 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d (a duplicate name collapsed an entry)",
			len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	// Spot-check both directions, including a post-gap opnum.
	if OpnumToName[OpnumRpcOpenPrinter] != "RpcOpenPrinter" {
		t.Errorf("OpnumToName[1] = %q, want RpcOpenPrinter", OpnumToName[OpnumRpcOpenPrinter])
	}
	if NameToOpnum["RpcXcvData"] != OpnumRpcXcvData || OpnumRpcXcvData != 88 {
		t.Errorf("NameToOpnum[RpcXcvData] = %d, want 88", NameToOpnum["RpcXcvData"])
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// 12345678-1234-abcd-ef00-0123456789ab, version 1.0.
	if id.UUID.A != 0x12345678 || id.UUID.B != 0x1234 || id.UUID.C != 0xabcd ||
		id.UUID.D != 0xef00 || id.UUID.E != 0x0123456789ab {
		t.Errorf("SyntaxID UUID = %+v, want 12345678-1234-abcd-ef00-0123456789ab", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}
