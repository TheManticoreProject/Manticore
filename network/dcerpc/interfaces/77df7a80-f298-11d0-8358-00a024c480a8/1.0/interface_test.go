package rpcinterface_77df7a80f29811d0835800a024c480a8_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the dscomm interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "77df7a80-f298-11d0-8358-00a024c480a8" {
		t.Fatalf("UUID = %s, want 77df7a80-f298-11d0-8358-00a024c480a8", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies representative opnums (including the gaps left by the "not used on
// the wire" and [callback] methods) and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumS_DSCreateObject != 0 || OpnumS_DSLookupEnd != 8 || OpnumS_DSDeleteObjectGuid != 10 {
		t.Fatalf("opnums = %d/%d/%d, want 0/8/10", OpnumS_DSCreateObject, OpnumS_DSLookupEnd, OpnumS_DSDeleteObjectGuid)
	}
	if OpnumS_DSQMSetMachineProperties != 19 || OpnumS_DSValidateServer != 22 || OpnumS_DSGetServerPort != 27 {
		t.Fatalf("opnums = %d/%d/%d, want 19/22/27", OpnumS_DSQMSetMachineProperties, OpnumS_DSValidateServer, OpnumS_DSGetServerPort)
	}
	if OpnumToName[0] != "S_DSCreateObject" || NameToOpnum["S_DSGetServerPort"] != 27 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 20 {
		t.Fatalf("on-the-wire method count = %d, want 20", len(OpnumToName))
	}
}

// TestSuccessResolvesThroughHRESULT pins the one status value this package no longer
// declares. MQ_OK is 0x00000000, which is S_OK, so the methods compare the retval against
// hresult.S_OK and StatusString renders it out of the shared table.
func TestSuccessResolvesThroughHRESULT(t *testing.T) {
	if uint32(hresult.S_OK) != 0x00000000 {
		t.Fatalf("hresult.S_OK = 0x%08x, want 0x00000000", uint32(hresult.S_OK))
	}
	if got := hresult.S_OK.String(); got != "S_OK" {
		t.Errorf("hresult.S_OK.String() = %q, want S_OK", got)
	}
	if resolved, defined := hresult.FromName("S_OK"); !defined || resolved != hresult.S_OK {
		t.Errorf("hresult.FromName(\"S_OK\") = 0x%08x, %v; want 0x00000000, true", uint32(resolved), defined)
	}
	if !hresult.S_OK.IsSuccess() || hresult.S_OK.Error() != nil {
		t.Errorf("hresult.S_OK reports failure; MQ_OK is a success status")
	}
	if got := StatusString(uint32(hresult.S_OK)); got != "S_OK" {
		t.Errorf("StatusString(0x00000000) = %q, want S_OK", got)
	}
}

// TestStatusStringDecodesMSMQFacility pins the reason the MQ_* block stays declared in this
// package: every one of these values sits in FACILITY_MSMQ (0x00E), a facility [MS-ERREF]
// 2.1.1 does not carry at all, so the shared table resolves none of them. StatusString
// names them and hresult.Lookup reports nothing for them; should a later pass route them
// through the shared table, this fails instead of silently renaming a directory error.
func TestStatusStringDecodesMSMQFacility(t *testing.T) {
	msmq := map[uint32]string{
		MQ_ERROR:                            "MQ_ERROR",
		MQ_ERROR_PROPERTY:                   "MQ_ERROR_PROPERTY",
		MQ_ERROR_QUEUE_NOT_FOUND:            "MQ_ERROR_QUEUE_NOT_FOUND",
		MQ_ERROR_QUEUE_EXISTS:               "MQ_ERROR_QUEUE_EXISTS",
		MQ_ERROR_INVALID_PARAMETER:          "MQ_ERROR_INVALID_PARAMETER",
		MQ_ERROR_INVALID_HANDLE:             "MQ_ERROR_INVALID_HANDLE",
		MQ_ERROR_NO_DS:                      "MQ_ERROR_NO_DS",
		MQ_ERROR_ILLEGAL_QUEUE_PATHNAME:     "MQ_ERROR_ILLEGAL_QUEUE_PATHNAME",
		MQ_ERROR_ACCESS_DENIED:              "MQ_ERROR_ACCESS_DENIED",
		MQ_ERROR_ILLEGAL_MQCOLUMNS:          "MQ_ERROR_ILLEGAL_MQCOLUMNS",
		MQ_ERROR_ILLEGAL_PROPID:             "MQ_ERROR_ILLEGAL_PROPID",
		MQ_ERROR_ILLEGAL_RELATION:           "MQ_ERROR_ILLEGAL_RELATION",
		MQ_ERROR_ILLEGAL_PROPERTY_VALUE:     "MQ_ERROR_ILLEGAL_PROPERTY_VALUE",
		MQ_ERROR_ILLEGAL_RESTRICTION_PROPID: "MQ_ERROR_ILLEGAL_RESTRICTION_PROPID",
		MQ_ERROR_DS_ERROR:                   "MQ_ERROR_DS_ERROR",
		MQ_ERROR_ILLEGAL_SORT_PROPID:        "MQ_ERROR_ILLEGAL_SORT_PROPID",
	}
	if len(msmq) != 16 {
		t.Fatalf("MSMQ status table has %d entries, want 16", len(msmq))
	}
	for code, name := range msmq {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if facility := (code >> 16) & 0x07FF; facility != 0x00E {
			t.Errorf("%s = 0x%08x, facility 0x%03x, want FACILITY_MSMQ 0x00E", name, code, facility)
		}
		if entry, defined := hresult.Lookup(hresult.HRESULT(code)); defined {
			t.Errorf("hresult.Lookup(0x%08x) resolves to %q; [MS-ERREF] 2.1.1 carries no FACILITY_MSMQ row", code, entry.Name)
		}
		if _, defined := hresult.FromName(name); defined {
			t.Errorf("hresult.FromName(%q) resolves; the code is declared in this package instead", name)
		}
		if hresult.HRESULT(code).IsSuccess() {
			t.Errorf("%s = 0x%08x reports success; the MQ_ERROR_* values are failures", name, code)
		}
	}
}

// TestStatusStringFallsThroughToHRESULT shows what the fall-through buys. A generic HRESULT
// the old switch never enumerated reached the caller as bare hex and now resolves by name,
// including one the shared table derives through HRESULT_FROM_WIN32, while a value the
// specification defines nowhere still renders as hex.
func TestStatusStringFallsThroughToHRESULT(t *testing.T) {
	generic := map[uint32]string{
		0x80004005: "E_FAIL",
		0x80070005: "E_ACCESSDENIED",
		0x8007000E: "E_OUTOFMEMORY",
	}
	for code, name := range generic {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}

	// A FACILITY_WIN32 code the table does not name is derived from the Win32 table.
	wrapped := hresult.FromWin32(win32.ERROR_INVALID_NAME)
	if uint32(wrapped) != 0x8007007B {
		t.Fatalf("hresult.FromWin32(ERROR_INVALID_NAME) = 0x%08x, want 0x8007007b", uint32(wrapped))
	}
	if got := StatusString(uint32(wrapped)); got != "HRESULT_FROM_WIN32(ERROR_INVALID_NAME)" {
		t.Errorf("StatusString(0x8007007b) = %q, want HRESULT_FROM_WIN32(ERROR_INVALID_NAME)", got)
	}
	if code, wraps := wrapped.ToWin32(); !wraps || code != win32.ERROR_INVALID_NAME {
		t.Errorf("ToWin32() = 0x%08x, %v; want 0x0000007b, true", uint32(code), wraps)
	}

	// A FACILITY_MSMQ value outside the list this package declares still renders as hex:
	// the shared table has no row for the facility to fall back on.
	if got := StatusString(0xC00E7FFF); got != "0xc00e7fff" {
		t.Errorf("StatusString(0xc00e7fff) = %q, want 0xc00e7fff", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want 0xdeadbeef", got)
	}
}
