package rpcinterface_88143fd0c28d4b2b8fef8d882f6a9390_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
)

// TestSyntaxID pins the abstract syntax identifier for the TermSrvEnumeration interface
// (88143fd0-c28d-4b2b-8fef-8d882f6a9390 v1.0, [MS-TSTS]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "88143fd0-c28d-4b2b-8fef-8d882f6a9390" {
		t.Errorf("UUID = %s, want 88143fd0-c28d-4b2b-8fef-8d882f6a9390", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) == 0 {
		t.Fatal("OpnumToName is empty")
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, n := range OpnumToName {
		if NameToOpnum[n] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, n, NameToOpnum[n])
		}
	}
}

// TestStatusCodesResolveThroughHRESULT pins that the HRESULT values this interface used
// to declare resolve through the shared [MS-ERREF] 2.1.1 table under their specification
// names, that a value the interface never enumerated now renders by name rather than as
// undecoded hex, that success is a range and not a single value, and that a value the
// specification does not define still renders as hex.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "S_OK",
		0x80004004: "E_ABORT",
		0x80004005: "E_FAIL",
		0x80070005: "E_ACCESSDENIED",
		0x8007000E: "E_OUTOFMEMORY",
		0x80070057: "E_INVALIDARG",
	}
	for code, name := range documented {
		if got := hresult.HRESULT(code).String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := hresult.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("hresult.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// E_POINTER was outside the subset this interface used to declare, so it rendered as
	// hex; it resolves by name now.
	if got := hresult.HRESULT(0x80004003).String(); got != "E_POINTER" {
		t.Errorf("hresult.HRESULT(0x80004003).String() = %q, want E_POINTER", got)
	}

	// Success is a range: S_FALSE has bit 31 clear, so it succeeds and reports no error.
	if !hresult.S_FALSE.IsSuccess() || hresult.S_FALSE.Error() != nil {
		t.Errorf("hresult.S_FALSE.IsSuccess() = %v, Error() = %v; want true, nil",
			hresult.S_FALSE.IsSuccess(), hresult.S_FALSE.Error())
	}

	// A value [MS-ERREF] 2.1.1 does not define still renders as hex.
	if got := hresult.HRESULT(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("hresult.HRESULT(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestPipeName pins the [MS-TSTS] section 1.9 endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\LSM_API_service` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\LSM_API_service`)
	}
}
