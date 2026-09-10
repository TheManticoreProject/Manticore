package rpcinterface_2f5f6521ca471068b31900dd010662db_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID checks the abstract syntax matches [MS-TRP] Appendix A.1 (remotesp):
// 2f5f6521-ca47-1068-b319-00dd010662db, version 1.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "2f5f6521-ca47-1068-b319-00dd010662db" {
		t.Errorf("UUID = %s, want 2f5f6521-ca47-1068-b319-00dd010662db", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent inverses and
// cover the three on-the-wire opnums.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 3 {
		t.Fatalf("OpnumToName has %d entries, want 3", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if OpnumRemoteSPAttach != 0 || OpnumRemoteSPEventProc != 1 || OpnumRemoteSPDetach != 2 {
		t.Errorf("opnums = %d/%d/%d, want 0/1/2", OpnumRemoteSPAttach, OpnumRemoteSPEventProc, OpnumRemoteSPDetach)
	}
}

// TestStatusCodesResolveThroughWin32 pins that the one code this descriptor used to
// declare for RemoteSPAttach's return resolves through the shared [MS-ERREF] 2.2 table
// under its specification name, that Win32 failures the one-entry subset could not name now
// render by name rather than as undecoded hex, that a TAPI value from outside that table
// still renders as hex, and that a value the specification does not define does too.
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

	// RemoteSPAttach's failures are "as specified in [MS-ERREF]", and success was the only
	// value the descriptor could name, so every one of them used to render as undecoded
	// hex. The reverse binding is established by the telephony server calling into the
	// client, so an RPC-layer failure is as likely here as a local one; all five resolve by
	// name now.
	outsideOldSubset := map[uint32]string{
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x0000000E: "ERROR_OUTOFMEMORY",
		0x000006BB: "RPC_S_SERVER_TOO_BUSY",
		0x000006BE: "RPC_S_CALL_FAILED",
	}
	for code, name := range outsideOldSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
	}

	// LINEERR_OPERATIONFAILED, the TAPI code [MS-TRP] names for the neighbouring
	// ClientAttach return, is in the 0x8000xxxx block [MS-ERREF] 2.2 does not cover. It has
	// no row, so it stays hexadecimal and cannot be misnamed out of the Win32 table.
	if _, defined := win32.Lookup(win32.WIN32_ERROR(0x80000048)); defined {
		t.Error("win32 names 0x80000048; it is a TAPI value and the Win32 table must not claim it")
	}
	if got := win32.WIN32_ERROR(0x80000048).String(); got != "0x80000048" {
		t.Errorf("win32.WIN32_ERROR(0x80000048).String() = %q, want hex", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xdeadbeef).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestPipeName documents that remotesp has no fixed named pipe: [MS-TRP] 2.1 uses a
// client-specified endpoint for the reverse callback connection, so PipeName is empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (client-specified reverse endpoint)", PipeName)
	}
}
