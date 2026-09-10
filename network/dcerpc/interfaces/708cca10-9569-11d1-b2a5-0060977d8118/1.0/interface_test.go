package rpcinterface_708cca10956911d1b2a50060977d8118_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the dscomm2 interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "708cca10-9569-11d1-b2a5-0060977d8118" {
		t.Fatalf("UUID = %s, want 708cca10-9569-11d1-b2a5-0060977d8118", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the implemented opnums (opnum 7 is not used on the wire) and the
// name mapping.
func TestOpnums(t *testing.T) {
	if OpnumS_DSGetComputerSites != 0 || OpnumS_DSIsServerGC != 6 || OpnumS_DSGetGCListInDomain != 8 {
		t.Fatalf("opnums = %d/%d/%d, want 0/6/8", OpnumS_DSGetComputerSites, OpnumS_DSIsServerGC, OpnumS_DSGetGCListInDomain)
	}
	if OpnumToName[0] != "S_DSGetComputerSites" || NameToOpnum["S_DSGetGCListInDomain"] != 8 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if _, ok := OpnumToName[7]; ok {
		t.Fatal("opnum 7 is not used on the wire and must not be mapped")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 8 {
		t.Fatalf("on-the-wire method count = %d, want 8", len(OpnumToName))
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
		MQ_ERROR:                        "MQ_ERROR",
		MQ_ERROR_PROPERTY:               "MQ_ERROR_PROPERTY",
		MQ_ERROR_QUEUE_NOT_FOUND:        "MQ_ERROR_QUEUE_NOT_FOUND",
		MQ_ERROR_INVALID_PARAMETER:      "MQ_ERROR_INVALID_PARAMETER",
		MQ_ERROR_INVALID_HANDLE:         "MQ_ERROR_INVALID_HANDLE",
		MQ_ERROR_NO_DS:                  "MQ_ERROR_NO_DS",
		MQ_ERROR_ILLEGAL_QUEUE_PATHNAME: "MQ_ERROR_ILLEGAL_QUEUE_PATHNAME",
		MQ_ERROR_ACCESS_DENIED:          "MQ_ERROR_ACCESS_DENIED",
		MQ_ERROR_ILLEGAL_PROPID:         "MQ_ERROR_ILLEGAL_PROPID",
		MQ_ERROR_ILLEGAL_PROPERTY_VALUE: "MQ_ERROR_ILLEGAL_PROPERTY_VALUE",
		MQ_ERROR_DS_ERROR:               "MQ_ERROR_DS_ERROR",
	}
	if len(retained) != 11 {
		t.Fatalf("retained table has %d entries, want 11", len(retained))
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
