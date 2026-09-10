package rpcinterface_1257b580ce2f410982d6a9459d0bf6bc_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the SessEnvPublicRpc interface
// (1257b580-ce2f-4109-82d6-a9459d0bf6bc v1.0, [MS-TSTS]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "1257b580-ce2f-4109-82d6-a9459d0bf6bc" {
		t.Errorf("UUID = %s, want 1257b580-ce2f-4109-82d6-a9459d0bf6bc", got)
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

// TestStatusCodesResolveThroughHRESULT pins that the HRESULT values [MS-TSTS] 3.10.4.1
// reports resolve through the shared [MS-ERREF] 2.1 table under their specification
// names, that a value this interface never enumerated now renders by name rather than
// as undecoded hex, and that a value the specification does not define still renders as
// hex.
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

	// A shadow request the target refuses fails with ERROR_CTX_SHADOW_DENIED. That
	// value is FACILITY_WIN32, so it is HRESULT_FROM_WIN32 of the Win32 code
	// ([MS-ERREF] 2.1.2). It was outside the subset this interface used to declare,
	// so it rendered as hex; it resolves by name now.
	if got := hresult.FromWin32(win32.ERROR_CTX_SHADOW_DENIED); uint32(got) != 0x80071B84 {
		t.Errorf("hresult.FromWin32(ERROR_CTX_SHADOW_DENIED) = 0x%08x, want 0x80071B84", uint32(got))
	}
	if got := hresult.HRESULT(0x80071B84).String(); got != "HRESULT_FROM_WIN32(ERROR_CTX_SHADOW_DENIED)" {
		t.Errorf("hresult.HRESULT(0x80071b84).String() = %q, want HRESULT_FROM_WIN32(ERROR_CTX_SHADOW_DENIED)", got)
	}

	// A value [MS-ERREF] 2.1 does not define still renders as hex.
	if got := hresult.HRESULT(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("hresult.HRESULT(0xdeadbeef).String() = %q, want hex", got)
	}
}

// TestPipeName pins the [MS-TSTS] section 1.9 endpoint.
func TestPipeName(t *testing.T) {
	if PipeName != `\SessEnvPublicRpc` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\SessEnvPublicRpc`)
	}
}
