package rpcinterface_0b6edbfa4a244fc68a23942b1eca65d1_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID checks the abstract syntax UUID and version match [MS-PAN] Appendix A.1.
func TestSyntaxID(t *testing.T) {
	sid := SyntaxID()
	if got := sid.UUID.ToFormatD(); got != "0b6edbfa-4a24-4fc6-8a23-942b1eca65d1" {
		t.Errorf("UUID = %s, want 0b6edbfa-4a24-4fc6-8a23-942b1eca65d1", got)
	}
	if sid.MajorVersion != 1 || sid.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", sid.MajorVersion, sid.MinorVersion)
	}
}

// TestPipeName pins the (empty) transport endpoint: MS-PAN is ncacn_ip_tcp only.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}

// TestOpnums verifies the opnums, that Opnum2NotUsedOnWire is omitted, and the name map.
func TestOpnums(t *testing.T) {
	if OpnumIRPCAsyncNotify_RegisterClient != 0 || OpnumIRPCAsyncNotify_GetNewChannel != 3 || OpnumIRPCAsyncNotify_CloseChannel != 6 {
		t.Fatalf("opnums = %d/%d/%d, want 0/3/6", OpnumIRPCAsyncNotify_RegisterClient, OpnumIRPCAsyncNotify_GetNewChannel, OpnumIRPCAsyncNotify_CloseChannel)
	}
	// Opnum 2 (Opnum2NotUsedOnWire) MUST NOT appear in the map.
	if _, ok := OpnumToName[2]; ok {
		t.Error("opnum 2 (Opnum2NotUsedOnWire) must be omitted from OpnumToName")
	}
	if len(OpnumToName) != 6 {
		t.Fatalf("on-the-wire method count = %d, want 6", len(OpnumToName))
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent inverses.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName has %d entries, NameToOpnum has %d", len(OpnumToName), len(NameToOpnum))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d (ok=%v), want %d", name, got, ok, op)
		}
	}
}

// TestStatusCodesResolveThroughHRESULT pins that the five HRESULTs this interface used to
// declare resolve through the shared [MS-ERREF] 2.1.1 table under the names the
// specification gives them, including the two the table derives from the Win32 code their
// FACILITY_WIN32 half carries, and that the table names failures the old subset could
// only render as hex.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	documented := map[hresult.HRESULT]string{
		0x00000000: "S_OK",
		0x80070005: "E_ACCESSDENIED",
		0x8007000E: "E_OUTOFMEMORY",
		0x80070015: "HRESULT_FROM_WIN32(ERROR_NOT_READY)",
		0x8007007B: "HRESULT_FROM_WIN32(ERROR_INVALID_NAME)",
	}
	for code, name := range documented {
		if got := code.String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", uint32(code), got, name)
		}
	}

	// The two FACILITY_WIN32 wrappings need no declaration of their own: they are built
	// from, and unwrap back to, the Win32 codes the specification names for them.
	if got := hresult.FromWin32(win32.ERROR_NOT_READY); got != 0x80070015 {
		t.Errorf("hresult.FromWin32(win32.ERROR_NOT_READY) = 0x%08x, want 0x80070015", uint32(got))
	}
	if got := hresult.FromWin32(win32.ERROR_INVALID_NAME); got != 0x8007007B {
		t.Errorf("hresult.FromWin32(win32.ERROR_INVALID_NAME) = 0x%08x, want 0x8007007b", uint32(got))
	}
	if code, wraps := hresult.HRESULT(0x80070015).ToWin32(); !wraps || code != win32.ERROR_NOT_READY {
		t.Errorf("hresult.HRESULT(0x80070015).ToWin32() = 0x%08x, %v; want 0x00000015, true", uint32(code), wraps)
	}

	// E_FAIL is a common [MS-ERREF] HRESULT a server may return under [MS-PAN] 3.1.4 and
	// the old subset did not declare, so it rendered as hex; it resolves by name now.
	if got := hresult.HRESULT(0x80004005).String(); got != "E_FAIL" {
		t.Errorf("hresult.HRESULT(0x80004005).String() = %q, want E_FAIL", got)
	}

	// A value [MS-ERREF] 2.1.1 does not define, and whose facility is not FACILITY_WIN32
	// either, still renders as hex.
	if got := hresult.HRESULT(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("hresult.HRESULT(0xdeadbeef).String() = %q, want hex", got)
	}
}
