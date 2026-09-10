package rpcinterface_f5cc5a184264101a8c5908002b2f8426_56_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the nspi interface
// (f5cc5a18-4264-101a-8c59-08002b2f8426 v56.0, [MS-NSPI]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "f5cc5a18-4264-101a-8c59-08002b2f8426" {
		t.Errorf("UUID = %s, want f5cc5a18-4264-101a-8c59-08002b2f8426", got)
	}
	if s.MajorVersion != 56 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 56.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// all 20 on-the-wire opnums are covered. Opnum 15 (Opnum15NotUsedOnWire) is intentionally
// absent, so the covered set is {0..14, 16..20}.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 20 {
		t.Fatalf("OpnumToName has %d entries, want 20", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for op := uint16(0); op <= 20; op++ {
		_, ok := OpnumToName[op]
		if op == 15 {
			if ok {
				t.Errorf("opnum 15 is not used on the wire and must be absent")
			}
			continue
		}
		if !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestGenericStatusCodesResolveThroughHRESULT pins the five permitted values of [MS-NSPI]
// section 2.2.2 that are generic HRESULTs to their [MS-ERREF] 2.1 names through the shared
// table in both directions, and shows StatusString reaching them by falling through rather
// than by a local case. The three FACILITY_WIN32 values are also checked against the
// HRESULT_FROM_WIN32 macro of [MS-ERREF] 2.1.2 they are built from.
func TestGenericStatusCodesResolveThroughHRESULT(t *testing.T) {
	generic := map[uint32]string{
		0x00000000: "S_OK",           // Success
		0x80004005: "E_FAIL",         // GeneralFailure (MAPI_E_CALL_FAILED)
		0x80070005: "E_ACCESSDENIED", // AccessDenied (MAPI_E_NO_ACCESS)
		0x8007000E: "E_OUTOFMEMORY",  // NotEnoughMemory (MAPI_E_NOT_ENOUGH_MEMORY)
		0x80070057: "E_INVALIDARG",   // InvalidParameter (MAPI_E_INVALID_PARAMETER)
	}
	for code, name := range generic {
		if got := hresult.HRESULT(code).String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := hresult.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("hresult.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
	}

	// The three 0x8007xxxx values are HRESULT_FROM_WIN32 wrappings ([MS-ERREF] 2.1.2), so
	// they build from and unwrap to their Win32 codes. The wrapped code for 0x8007000E is
	// ERROR_OUTOFMEMORY (0x0000000E) and not the ERROR_NOT_ENOUGH_MEMORY (0x00000008) the
	// value's former private name suggested: 0x00000008 wraps to 0x80070008, a different
	// HRESULT altogether.
	wrapped := map[uint32]win32.WIN32_ERROR{
		0x80070005: win32.ERROR_ACCESS_DENIED,
		0x8007000E: win32.ERROR_OUTOFMEMORY,
		0x80070057: win32.ERROR_INVALID_PARAMETER,
	}
	for code, want := range wrapped {
		if got := hresult.FromWin32(want); uint32(got) != code {
			t.Errorf("hresult.FromWin32(0x%08x) = 0x%08x, want 0x%08x", uint32(want), uint32(got), code)
		}
		if got, wraps := hresult.HRESULT(code).ToWin32(); !wraps || got != want {
			t.Errorf("hresult.HRESULT(0x%08x).ToWin32() = 0x%08x, %v; want 0x%08x, true", code, uint32(got), wraps, uint32(want))
		}
	}

	// A value neither [MS-NSPI] nor [MS-ERREF] 2.1 defines still renders as hex.
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(0xdeadbeef) = %q, want 0xdeadbeef", got)
	}
}

// TestStatusStringDecodesMAPIFacilityITFCodes pins the one reason StatusString still
// exists: the twelve MAPI values this interface declares live in FACILITY_ITF (0x004),
// the facility [MS-ERREF] 2.1 reserves for interface-specific meanings, and the shared
// table either has no row for them or names the same value for an unrelated OLE error.
// Each must keep rendering under its MAPI name here, and the two colliding values must
// keep reporting the table's different name through hresult.Lookup, so that a later pass
// folding them into the shared table fails loudly instead of silently renaming a failed
// address-book logon.
func TestStatusStringDecodesMAPIFacilityITFCodes(t *testing.T) {
	mapiOnly := map[uint32]string{
		StatusErrorsReturned:     "MAPI_W_ERRORS_RETURNED",
		StatusNotSupported:       "MAPI_E_NO_SUPPORT",
		StatusInvalidObject:      "MAPI_E_INVALID_OBJECT",
		StatusOutOfResources:     "MAPI_E_NOT_ENOUGH_RESOURCES",
		StatusNotFound:           "MAPI_E_NOT_FOUND",
		StatusLogonFailed:        "MAPI_E_LOGON_FAILED",
		StatusTooComplex:         "MAPI_E_TOO_COMPLEX",
		StatusInvalidCodepage:    "MAPI_E_UNKNOWN_CPID",
		StatusInvalidLocale:      "MAPI_E_UNKNOWN_LCID",
		StatusTableTooBig:        "MAPI_E_TABLE_TOO_BIG",
		StatusInvalidBookmark:    "MAPI_E_INVALID_BOOKMARK",
		StatusAmbiguousRecipient: "MAPI_E_AMBIGUOUS_RECIP",
	}
	if len(mapiOnly) != 12 {
		t.Fatalf("MAPI-only table has %d entries, want 12", len(mapiOnly))
	}
	for code, name := range mapiOnly {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if (code>>16)&0x07FF != 0x004 {
			t.Errorf("%s = 0x%08x, facility %#x, want FACILITY_ITF (0x004)", name, code, (code>>16)&0x07FF)
		}
		if _, defined := hresult.FromName(name); defined {
			t.Errorf("hresult.FromName(%q) resolves; the code belongs in the shared table instead", name)
		}
	}

	// The collision that makes the local block necessary: for these two values the shared
	// table holds a row, and it names something unrelated.
	colliding := map[uint32]string{
		StatusNotSupported: "DRAGDROP_E_INVALIDHWND",
		StatusLogonFailed:  "CLASS_E_CLASSNOTAVAILABLE",
	}
	for code, tableName := range colliding {
		entry, defined := hresult.Lookup(hresult.HRESULT(code))
		if !defined || entry.Name != tableName {
			t.Errorf("hresult.Lookup(0x%08x) = %q, %v; want %q, true", code, entry.Name, defined, tableName)
		}
		if entry.Name == StatusString(code) {
			t.Errorf("0x%08x: shared table and StatusString agree on %q; the collision this test pins is gone", code, entry.Name)
		}
	}

	// The other ten have no [MS-ERREF] 2.1.1 row at all, so the shared table would render
	// them as bare hex where StatusString names them.
	for code := range mapiOnly {
		if _, colliding := colliding[code]; colliding {
			continue
		}
		if entry, defined := hresult.Lookup(hresult.HRESULT(code)); defined {
			t.Errorf("hresult.Lookup(0x%08x) = %q, true; want no entry", code, entry.Name)
		}
	}
}

// TestErrorsReturnedIsSuccessSeverity records that the ErrorsReturned warning of
// [MS-OXCDATA] section 2.5 has the severity bit clear and is therefore a success by the
// HRESULT rule of [MS-ERREF] 2.1, while the seven methods that tolerate it name it
// explicitly alongside S_OK and the other thirteen do not. Which statuses a method
// tolerates is [MS-NSPI]'s to decide, so the stubs keep the shape they had; this pins the
// severity fact so the discrepancy stays visible.
func TestErrorsReturnedIsSuccessSeverity(t *testing.T) {
	if !hresult.HRESULT(StatusErrorsReturned).IsSuccess() {
		t.Errorf("hresult.HRESULT(0x%08x).IsSuccess() = false, want true", StatusErrorsReturned)
	}
	if err := hresult.HRESULT(StatusErrorsReturned).Error(); err != nil {
		t.Errorf("hresult.HRESULT(0x%08x).Error() = %v, want nil", StatusErrorsReturned, err)
	}
	if !hresult.S_OK.IsSuccess() {
		t.Error("hresult.S_OK.IsSuccess() = false, want true")
	}
	for _, code := range []uint32{StatusNotSupported, StatusLogonFailed, 0x80004005} {
		if hresult.HRESULT(code).IsSuccess() {
			t.Errorf("hresult.HRESULT(0x%08x).IsSuccess() = true, want false", code)
		}
	}
}

// TestPipeName pins NSPI's empty pipe: it has no well-known named-pipe endpoint
// ([MS-NSPI] 2.1 uses dynamic endpoints across ncacn_np/ncacn_http/ncacn_ip_tcp).
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty", PipeName)
	}
}
