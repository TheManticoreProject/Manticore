package rpcinterface_22e5386d8b124bf0b0ec6a1ea419e366_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "22e5386d-8b12-4bf0-b0ec-6a1ea419e366" {
		t.Fatalf("UUID = %s, want 22e5386d-8b12-4bf0-b0ec-6a1ea419e366", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the implemented opnums and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumRpcNetEventOpenSession != 0 || OpnumRpcNetEventReceiveData != 1 || OpnumRpcNetEventCloseSession != 2 {
		t.Fatalf("opnums = %d/%d/%d, want 0/1/2", OpnumRpcNetEventOpenSession, OpnumRpcNetEventReceiveData, OpnumRpcNetEventCloseSession)
	}
	if OpnumToName[0] != "RpcNetEventOpenSession" || NameToOpnum["RpcNetEventCloseSession"] != 2 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughWin32 pins that the one Win32 code this descriptor used to
// declare — ERROR_SUCCESS, the value [MS-LREC] 3.1.4.2 documents for a successful DWORD
// return — resolves through the shared [MS-ERREF] 2.2 table under its specification name,
// that failure codes the one-entry subset could not name now render by name rather than as
// undecoded hex, and that a value the specification does not define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
	}
	if len(documented) != 1 {
		t.Fatalf("documented table has %d entries, want the 1 code the descriptor declared", len(documented))
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// [MS-LREC] 3.1.4.2 enumerates no failure codes, so success was the only value the
	// descriptor could name and every failure the event forwarder reports used to render
	// as undecoded hex. These are the kind of codes an open or a drain fails with — the
	// caller lacks access, the logger name does not match a started session, the context
	// handle is stale — and each one resolves by name now.
	outsideOldSubset := map[uint32]string{
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x00000490: "ERROR_NOT_FOUND",
		0x00001069: "ERROR_WMI_INSTANCE_NOT_FOUND",
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
