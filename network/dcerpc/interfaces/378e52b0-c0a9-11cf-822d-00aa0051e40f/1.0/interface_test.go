package rpcinterface_378e52b0c0a911cf822d00aa0051e40f_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of sasec.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "378e52b0-c0a9-11cf-822d-00aa0051e40f" {
		t.Fatalf("UUID = %s, want 378e52b0-c0a9-11cf-822d-00aa0051e40f", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName verifies the well-known ncacn_np endpoint shared with ATSvc.
func TestPipeName(t *testing.T) {
	if PipeName != `\atsvc` {
		t.Fatalf("PipeName = %q, want \\atsvc", PipeName)
	}
}

// TestOpnums verifies the implemented opnums and the name mapping round trip.
func TestOpnums(t *testing.T) {
	if OpnumSASetAccountInformation != 0 || OpnumSASetNSAccountInformation != 1 ||
		OpnumSAGetNSAccountInformation != 2 || OpnumSAGetAccountInformation != 3 {
		t.Fatalf("opnums = %d/%d/%d/%d, want 0/1/2/3",
			OpnumSASetAccountInformation, OpnumSASetNSAccountInformation,
			OpnumSAGetNSAccountInformation, OpnumSAGetAccountInformation)
	}
	if OpnumToName[0] != "SASetAccountInformation" || NameToOpnum["SAGetAccountInformation"] != 3 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughHRESULT pins that the six HRESULTs this interface used to
// declare resolve through the shared [MS-ERREF] 2.1.1 table under the names the
// specification gives them, including the three the table derives from the Win32 code
// their FACILITY_WIN32 half carries, and that the table names failures the old subset
// could only render as hex.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	documented := map[hresult.HRESULT]string{
		0x00000000: "S_OK",
		0x80070002: "HRESULT_FROM_WIN32(ERROR_FILE_NOT_FOUND)",
		0x80070005: "E_ACCESSDENIED",
		0x8007000D: "HRESULT_FROM_WIN32(ERROR_INVALID_DATA)",
		0x80070057: "E_INVALIDARG",
		0x8007052E: "HRESULT_FROM_WIN32(ERROR_LOGON_FAILURE)",
	}
	for code, name := range documented {
		if got := code.String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", uint32(code), got, name)
		}
	}

	// The FACILITY_WIN32 wrappings need no declaration of their own: they are built from,
	// and unwrap back to, the Win32 codes [MS-TSCH] 3.2.5.3 names for them.
	if got := hresult.FromWin32(win32.ERROR_LOGON_FAILURE); got != 0x8007052E {
		t.Errorf("hresult.FromWin32(win32.ERROR_LOGON_FAILURE) = 0x%08x, want 0x8007052e", uint32(got))
	}
	if code, wraps := hresult.HRESULT(0x80070002).ToWin32(); !wraps || code != win32.ERROR_FILE_NOT_FOUND {
		t.Errorf("hresult.HRESULT(0x80070002).ToWin32() = 0x%08x, %v; want 0x00000002, true", uint32(code), wraps)
	}

	// ERROR_ACCOUNT_DISABLED wrapped as an HRESULT is a credential failure the SASet*
	// methods can report alongside the ERROR_LOGON_FAILURE the old subset declared, and it
	// was outside that subset, so it rendered as hex; it resolves by name now.
	if got := hresult.FromWin32(win32.ERROR_ACCOUNT_DISABLED); got != 0x80070533 {
		t.Errorf("hresult.FromWin32(win32.ERROR_ACCOUNT_DISABLED) = 0x%08x, want 0x80070533", uint32(got))
	}
	if got := hresult.HRESULT(0x80070533).String(); got != "HRESULT_FROM_WIN32(ERROR_ACCOUNT_DISABLED)" {
		t.Errorf("hresult.HRESULT(0x80070533).String() = %q, want HRESULT_FROM_WIN32(ERROR_ACCOUNT_DISABLED)", got)
	}

	// A value [MS-ERREF] 2.1.1 does not define, and whose facility is not FACILITY_WIN32
	// either, still renders as hex.
	if got := hresult.HRESULT(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("hresult.HRESULT(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestIsSuccess verifies the HRESULT success predicate, which is the severity bit and so
// a range rather than the single value zero ([MS-TSCH] 3.2.5.3, [MS-ERREF] 2.1).
func TestIsSuccess(t *testing.T) {
	if !IsSuccess(uint32(hresult.S_OK)) {
		t.Error("IsSuccess(S_OK) = false, want true")
	}
	if !IsSuccess(uint32(hresult.S_FALSE)) {
		t.Error("IsSuccess(S_FALSE) = false, want true: success is a severity bit, not zero alone")
	}
	if IsSuccess(0x80070005) {
		t.Error("IsSuccess(E_ACCESSDENIED) = true, want false")
	}
	if IsSuccess(0x8007052E) {
		t.Error("IsSuccess(HRESULT_FROM_WIN32(ERROR_LOGON_FAILURE)) = true, want false")
	}
}
