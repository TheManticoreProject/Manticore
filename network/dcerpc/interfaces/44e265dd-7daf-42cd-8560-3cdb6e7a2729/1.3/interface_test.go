package rpcinterface_44e265dd7daf42cd85603cdb6e7a2729_1_3

import (
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the TsProxyRpcInterface
// (44e265dd-7daf-42cd-8560-3cdb6e7a2729 v1.3, [MS-TSGU]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "44e265dd-7daf-42cd-8560-3cdb6e7a2729" {
		t.Errorf("UUID = %s, want 44e265dd-7daf-42cd-8560-3cdb6e7a2729", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 3 {
		t.Errorf("version = %d.%d, want 1.3", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses so the
// two maps never drift, and that only the 8 on-the-wire opnums are present (0 and 5 are
// "not used on the wire").
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 8 {
		t.Fatalf("OpnumToName has %d entries, want 8", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for _, op := range []uint16{1, 2, 3, 4, 6, 7, 8, 9} {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	// Opnums 0 and 5 are not used on the wire and must be absent.
	for _, op := range []uint16{0, 5} {
		if _, ok := OpnumToName[op]; ok {
			t.Errorf("opnum %d should be absent (NotUsedOnWire)", op)
		}
	}
}

// TestMigratedStatusCodesResolveThroughSharedTables pins the four values this
// descriptor no longer declares: SEC_E_LOGON_DENIED, which [MS-ERREF] 2.1.1 names, and
// the three Win32 codes of [MS-TSGU] 2.2.6, which [MS-ERREF] 2.2 names. It also pins
// that a value outside the old subset now renders by name, and that an undefined value
// still renders as hex.
func TestMigratedStatusCodesResolveThroughSharedTables(t *testing.T) {
	if got := uint32(hresult.SEC_E_LOGON_DENIED); got != 0x8009030C {
		t.Errorf("hresult.SEC_E_LOGON_DENIED = 0x%08x, want 0x8009030c", got)
	}
	if got := StatusString(uint32(hresult.SEC_E_LOGON_DENIED)); got != "SEC_E_LOGON_DENIED" {
		t.Errorf("StatusString(0x8009030c) = %q, want SEC_E_LOGON_DENIED", got)
	}

	win32Codes := map[win32.WIN32_ERROR]string{
		win32.ERROR_ACCESS_DENIED:       "ERROR_ACCESS_DENIED",
		win32.ERROR_BAD_ARGUMENTS:       "ERROR_BAD_ARGUMENTS",
		win32.ERROR_GRACEFUL_DISCONNECT: "ERROR_GRACEFUL_DISCONNECT",
	}
	for code, name := range win32Codes {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", uint32(code), got, name)
		}
		if got := StatusString(uint32(code)); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", uint32(code), got, name)
		}
		// Each is a Win32 code, not an HRESULT: read as one it is a
		// success-severity value the HRESULT table cannot name, which is why
		// StatusString routes these three through the Win32 table.
		if _, defined := hresult.Lookup(hresult.HRESULT(code)); defined {
			t.Errorf("hresult.Lookup(0x%08x) resolves; %s is a Win32 code, not an HRESULT", uint32(code), name)
		}
		if !hresult.HRESULT(code).IsSuccess() {
			t.Errorf("HRESULT(0x%08x).IsSuccess() = false, want true for a bare Win32 code", uint32(code))
		}
	}

	// E_FAIL was outside the subset this descriptor used to declare, so it rendered as
	// hex; it resolves by name now, through StatusString as well.
	if got := StatusString(0x80004005); got != "E_FAIL" {
		t.Errorf("StatusString(0x80004005) = %q, want E_FAIL", got)
	}
	// A value neither table defines still renders as hex.
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
}

// TestStatusStringDecodesGatewayPrivateCodes pins the reason StatusString still exists.
// The E_PROXY_* HRESULTs sit in FACILITY_WIN32 over the Win32 code range Terminal
// Services Gateway assigns for itself, which [MS-ERREF] names in neither table, and the
// E_PROXY_*_CODE values are HRESULT_CODE rather than HRESULT — success-severity if read
// as one. Both blocks must therefore keep decoding locally.
func TestStatusStringDecodesGatewayPrivateCodes(t *testing.T) {
	proxyHRESULTs := map[uint32]string{
		E_PROXY_INTERNALERROR:                       "E_PROXY_INTERNALERROR",
		E_PROXY_RAP_ACCESSDENIED:                    "E_PROXY_RAP_ACCESSDENIED",
		E_PROXY_NAP_ACCESSDENIED:                    "E_PROXY_NAP_ACCESSDENIED",
		E_PROXY_ALREADYDISCONNECTED:                 "E_PROXY_ALREADYDISCONNECTED",
		E_PROXY_CAPABILITYMISMATCH:                  "E_PROXY_CAPABILITYMISMATCH",
		E_PROXY_QUARANTINE_ACCESSDENIED:             "E_PROXY_QUARANTINE_ACCESSDENIED",
		E_PROXY_NOCERTAVAILABLE:                     "E_PROXY_NOCERTAVAILABLE",
		E_PROXY_COOKIE_BADPACKET:                    "E_PROXY_COOKIE_BADPACKET",
		E_PROXY_COOKIE_AUTHENTICATION_ACCESS_DENIED: "E_PROXY_COOKIE_AUTHENTICATION_ACCESS_DENIED",
		E_PROXY_UNSUPPORTED_AUTHENTICATION_METHOD:   "E_PROXY_UNSUPPORTED_AUTHENTICATION_METHOD",
	}
	if len(proxyHRESULTs) != 10 {
		t.Fatalf("gateway HRESULT table has %d entries, want 10", len(proxyHRESULTs))
	}
	for code, name := range proxyHRESULTs {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if _, defined := hresult.Lookup(hresult.HRESULT(code)); defined {
			t.Errorf("hresult.Lookup(0x%08x) resolves; %s belongs in the shared table instead", code, name)
		}
		if want := fmt.Sprintf("0x%08x", code); hresult.HRESULT(code).String() != want {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want the hex rendering %q", code, hresult.HRESULT(code).String(), want)
		}
		// The retention reason, derived from the value: FACILITY_WIN32, over a code
		// [MS-ERREF] 2.2 does not name either, so the HRESULT_FROM_WIN32 derivation
		// cannot resolve it.
		wrapped, wraps := hresult.HRESULT(code).ToWin32()
		if !wraps {
			t.Errorf("HRESULT(0x%08x).ToWin32() reports no Win32 code, want FACILITY_WIN32", code)
		}
		if wrapped < 0x59D8 || wrapped > 0x59F9 {
			t.Errorf("%s wraps 0x%04x, outside the gateway range 0x59D8..0x59F9", name, uint32(wrapped))
		}
		if _, defined := win32.Lookup(wrapped); defined {
			t.Errorf("win32.Lookup(0x%04x) resolves; %s could be built with hresult.FromWin32 instead", uint32(wrapped), name)
		}
	}

	proxyCodes := map[uint32]string{
		E_PROXY_CONNECTIONABORTED_CODE:       "E_PROXY_CONNECTIONABORTED",
		E_PROXY_INTERNALERROR_CODE:           "E_PROXY_INTERNALERROR (HRESULT_CODE)",
		E_PROXY_TS_CONNECTFAILED_CODE:        "E_PROXY_TS_CONNECTFAILED",
		E_PROXY_MAXCONNECTIONSREACHED_CODE:   "E_PROXY_MAXCONNECTIONSREACHED",
		E_PROXY_NOTSUPPORTED_CODE:            "E_PROXY_NOTSUPPORTED",
		E_PROXY_SESSIONTIMEOUT_CODE:          "E_PROXY_SESSIONTIMEOUT",
		E_PROXY_REAUTH_AUTHN_FAILED_CODE:     "E_PROXY_REAUTH_AUTHN_FAILED",
		E_PROXY_REAUTH_CAP_FAILED_CODE:       "E_PROXY_REAUTH_CAP_FAILED",
		E_PROXY_REAUTH_RAP_FAILED_CODE:       "E_PROXY_REAUTH_RAP_FAILED",
		E_PROXY_SDR_NOT_SUPPORTED_BY_TS_CODE: "E_PROXY_SDR_NOT_SUPPORTED_BY_TS",
		E_PROXY_REAUTH_NAP_FAILED_CODE:       "E_PROXY_REAUTH_NAP_FAILED",
	}
	if len(proxyCodes) != 11 {
		t.Fatalf("gateway HRESULT_CODE table has %d entries, want 11", len(proxyCodes))
	}
	for code, name := range proxyCodes {
		if got := StatusString(code); got != name {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, name)
		}
		if _, defined := hresult.Lookup(hresult.HRESULT(code)); defined {
			t.Errorf("hresult.Lookup(0x%08x) resolves; %s belongs in the shared table instead", code, name)
		}
		// Read as an HRESULT each of these is a success, which is exactly why the
		// stubs test the status against StatusSuccess and not against IsSuccess.
		if !hresult.HRESULT(code).IsSuccess() {
			t.Errorf("HRESULT(0x%08x).IsSuccess() = false, want true for an HRESULT_CODE value", code)
		}
	}

	// The one collision in that block: [MS-ERREF] 2.2 names 0x000004D4 for the generic
	// socket error, not for the gateway, so it cannot be read out of the shared table.
	if got := win32.WIN32_ERROR(E_PROXY_CONNECTIONABORTED_CODE).String(); got != "ERROR_CONNECTION_ABORTED" {
		t.Errorf("win32.WIN32_ERROR(0x000004d4).String() = %q, want ERROR_CONNECTION_ABORTED", got)
	}
}
