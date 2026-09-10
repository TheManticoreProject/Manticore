package rpcinterface_1a927394352e4553ae3f7cf4aafca620_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the WdsRpcInterface interface
// (1a927394-352e-4553-ae3f-7cf4aafca620 v1.0, [MS-WDSC]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "1a927394-352e-4553-ae3f-7cf4aafca620" {
		t.Errorf("UUID = %s, want 1a927394-352e-4553-ae3f-7cf4aafca620", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover the single opnum 0.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 1 {
		t.Fatalf("OpnumToName has %d entries, want 1", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	if OpnumToName[OpnumWdsRpcMessage] != "WdsRpcMessage" {
		t.Errorf("OpnumToName[%d] = %q, want WdsRpcMessage", OpnumWdsRpcMessage, OpnumToName[OpnumWdsRpcMessage])
	}
	for op, n := range OpnumToName {
		if NameToOpnum[n] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, n, NameToOpnum[n])
		}
	}
}

// TestStatusCodesResolveThroughWin32 pins that the five Win32 codes this descriptor used
// to declare for WdsRpcMessage ([MS-WDSC] 3.1.4.1) resolve through the shared [MS-ERREF]
// 2.2 table under their specification names, that codes outside that subset now render by
// name rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
	}
	if len(documented) != 5 {
		t.Fatalf("documented table has %d entries, want the 5 codes the descriptor declared", len(documented))
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// [MS-WDSC] 3.1.4.1 enumerates no fixed set of failures, and the WDS server fails a
	// request whose Endpoint Header, Endpoint GUID, OpCode or Variables Section is
	// invalid. These are the kind of codes such a failure carries; all four sat outside
	// the five-code subset and used to render as undecoded hex.
	outsideOldSubset := map[uint32]string{
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x00000490: "ERROR_NOT_FOUND",
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

// TestPipeName pins the descriptor-uniformity endpoint name. WdsRpcInterface is actually
// an ncacn_ip_tcp dynamic-endpoint interface ([MS-WDSC] 2.1); the constant is retained
// only for uniformity across interface descriptors.
func TestPipeName(t *testing.T) {
	if PipeName != `\WdsRpcInterface` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\WdsRpcInterface`)
	}
}
