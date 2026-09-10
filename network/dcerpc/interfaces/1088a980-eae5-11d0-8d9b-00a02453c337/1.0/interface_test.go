package rpcinterface_1088a980eae511d08d9b00a02453c337_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the qm2qm interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "1088a980-eae5-11d0-8d9b-00a02453c337" {
		t.Fatalf("UUID = %s, want 1088a980-eae5-11d0-8d9b-00a02453c337", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the (empty) transport endpoint: qm2qm is ncacn_ip_tcp only.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}

// TestOpnums verifies the opnums and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumRemoteQMStartReceive != 0 || OpnumRemoteQMGetQMQMServerPort != 7 || OpnumRemoteQMStartReceiveByLookupId != 10 {
		t.Fatalf("opnums = %d/%d/%d, want 0/7/10", OpnumRemoteQMStartReceive, OpnumRemoteQMGetQMQMServerPort, OpnumRemoteQMStartReceiveByLookupId)
	}
	if OpnumToName[0] != "RemoteQMStartReceive" || NameToOpnum["RemoteQmGetVersion"] != 8 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 11 {
		t.Fatalf("on-the-wire method count = %d, want 11", len(OpnumToName))
	}
}

// TestStatusStringKeepsMSMQFacilityAndDefersTheRest pins both halves of the split. MQ_OK
// left this package: it is 0x00000000, the value the shared HRESULT package declares as
// S_OK, so it resolves there and the stubs compare against that constant. The MQ_ERROR_*
// codes could not leave: they are HRESULTs in FACILITY_MSMQ (0x00E), a facility for which
// [MS-ERREF] section 2.1.1 carries no row, so hresult.Lookup reports nothing for any of
// them and StatusString must name them itself.
func TestStatusStringKeepsMSMQFacilityAndDefersTheRest(t *testing.T) {
	if hresult.S_OK != 0x00000000 {
		t.Fatalf("hresult.S_OK = 0x%08x, want 0x00000000 (the value MQ_OK had)", uint32(hresult.S_OK))
	}
	if got := hresult.S_OK.String(); got != "S_OK" {
		t.Errorf("hresult.S_OK.String() = %q, want S_OK", got)
	}
	if err := hresult.S_OK.Error(); err != nil {
		t.Errorf("hresult.S_OK.Error() = %v, want nil", err)
	}
	if got := StatusString(uint32(hresult.S_OK)); got != "S_OK" {
		t.Errorf("StatusString(0x00000000) = %q, want S_OK", got)
	}

	retained := map[uint32]string{
		MQ_ERROR:                   "MQ_ERROR",
		MQ_ERROR_INVALID_PARAMETER: "MQ_ERROR_INVALID_PARAMETER",
		MQ_ERROR_INVALID_HANDLE:    "MQ_ERROR_INVALID_HANDLE",
		MQ_ERROR_IO_TIMEOUT:        "MQ_ERROR_IO_TIMEOUT",
	}
	if len(retained) != 4 {
		t.Fatalf("retained table has %d entries, want 4", len(retained))
	}
	for code, name := range retained {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if facility := (code >> 16) & 0x07FF; facility != 0x00E {
			t.Errorf("%s = 0x%08x, facility 0x%03x, want FACILITY_MSMQ (0x00E)", name, code, facility)
		}
		if entry, defined := hresult.Lookup(hresult.HRESULT(code)); defined {
			t.Errorf("hresult.Lookup(0x%08x) resolves to %q; the code belongs in the shared table instead", code, entry.Name)
		}
		if hresult.HRESULT(code).IsSuccess() {
			t.Errorf("%s = 0x%08x reports success, want failure severity", name, code)
		}
	}

	// The fallthrough is the gain: a generic HRESULT this interface never enumerated used to
	// render as undecoded hex and now resolves by name.
	if got := StatusString(0x80004005); got != "E_FAIL" {
		t.Errorf("StatusString(0x80004005) = %q, want E_FAIL", got)
	}
	if got := StatusString(0x80070005); got != "E_ACCESSDENIED" {
		t.Errorf("StatusString(0x80070005) = %q, want E_ACCESSDENIED", got)
	}

	// So does a FACILITY_WIN32 value the specification's table does not name, which resolves
	// through the Win32 code it wraps.
	wrapped := hresult.FromWin32(win32.ERROR_FILE_NOT_FOUND)
	if uint32(wrapped) != 0x80070002 {
		t.Fatalf("hresult.FromWin32(ERROR_FILE_NOT_FOUND) = 0x%08x, want 0x80070002", uint32(wrapped))
	}
	if code, wraps := wrapped.ToWin32(); !wraps || code != win32.ERROR_FILE_NOT_FOUND {
		t.Errorf("ToWin32() = 0x%08x, %v; want 0x%08x, true", uint32(code), wraps, uint32(win32.ERROR_FILE_NOT_FOUND))
	}
	if got := StatusString(uint32(wrapped)); got != "HRESULT_FROM_WIN32(ERROR_FILE_NOT_FOUND)" {
		t.Errorf("StatusString(0x80070002) = %q, want HRESULT_FROM_WIN32(ERROR_FILE_NOT_FOUND)", got)
	}

	// A value neither table defines still renders as hex, as it always did.
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want 0xdeadbeef", got)
	}
}

// TestStatusStringKeepsTheCitedNTSTATUS pins the second reason StatusString still has a
// local table. STATUS_INVALID_PARAMETER is not an HRESULT: it is the NTSTATUS 0xC000000D
// that [MS-MQQP] cites directly, named by [MS-ERREF] section 2.3.1. Read as an HRESULT the
// same value is FACILITY_NULL code 0x000D, which [MS-ERREF] 2.1.1 does not define, so the
// shared HRESULT table cannot name it and the local arm has to.
func TestStatusStringKeepsTheCitedNTSTATUS(t *testing.T) {
	if STATUS_INVALID_PARAMETER != 0xC000000D {
		t.Fatalf("STATUS_INVALID_PARAMETER = 0x%08x, want 0xc000000d", STATUS_INVALID_PARAMETER)
	}
	if got := StatusString(STATUS_INVALID_PARAMETER); got != "STATUS_INVALID_PARAMETER" {
		t.Errorf("StatusString(0xc000000d) = %q, want STATUS_INVALID_PARAMETER", got)
	}
	if entry, defined := hresult.Lookup(hresult.HRESULT(STATUS_INVALID_PARAMETER)); defined {
		t.Errorf("hresult.Lookup(0xc000000d) resolves to %q; [MS-ERREF] 2.1.1 carries no such row", entry.Name)
	}
	if facility := (STATUS_INVALID_PARAMETER >> 16) & 0x07FF; facility != 0x000 {
		t.Errorf("as an HRESULT 0xc000000d has facility 0x%03x, want FACILITY_NULL (0x000)", facility)
	}
	// The NTSTATUS table is the one that carries the value, under its own NT_-prefixed
	// spelling of the name, so routing this arm there would change what StatusString renders.
	if got := nt_status.NT_STATUS(STATUS_INVALID_PARAMETER).String(); got != "NT_STATUS_INVALID_PARAMETER" {
		t.Errorf("nt_status.NT_STATUS(0xc000000d).String() = %q, want NT_STATUS_INVALID_PARAMETER", got)
	}
}
