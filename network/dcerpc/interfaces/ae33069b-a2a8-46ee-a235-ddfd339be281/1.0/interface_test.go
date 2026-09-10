package rpcinterface_ae33069ba2a846eea235ddfd339be281_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
)

// TestSyntaxID checks the abstract syntax UUID and version match [MS-PAN] Appendix A.2.
func TestSyntaxID(t *testing.T) {
	sid := SyntaxID()
	if got := sid.UUID.ToFormatD(); got != "ae33069b-a2a8-46ee-a235-ddfd339be281" {
		t.Errorf("UUID = %s, want ae33069b-a2a8-46ee-a235-ddfd339be281", got)
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

// TestStatusCodesResolveThroughHRESULT pins that the four HRESULTs this interface used to
// declare resolve through the shared [MS-ERREF] 2.1.1 table under the names the
// specification gives them, that the table names common HRESULTs the old subset could
// only render as hex, and that a value the specification does not define still renders as
// hex.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	documented := map[hresult.HRESULT]string{
		0x00000000: "S_OK",
		0x80070005: "E_ACCESSDENIED",
		0x8007000E: "E_OUTOFMEMORY",
		0x80070057: "E_INVALIDARG",
	}
	for code, name := range documented {
		if got := code.String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", uint32(code), got, name)
		}
		if resolved, defined := hresult.FromName(name); !defined || resolved != code {
			t.Errorf("hresult.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, uint32(code))
		}
	}

	// E_FAIL and E_UNEXPECTED are common [MS-ERREF] HRESULTs of the kind [MS-PAN] 3.1.2.4
	// allows and the old subset did not declare, so they rendered as hex; they resolve by
	// name now.
	for code, name := range map[hresult.HRESULT]string{0x80004005: "E_FAIL", 0x8000FFFF: "E_UNEXPECTED"} {
		if got := code.String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", uint32(code), got, name)
		}
	}

	// A value [MS-ERREF] 2.1.1 does not define, and whose facility is not FACILITY_WIN32
	// either, still renders as hex.
	if got := hresult.HRESULT(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("hresult.HRESULT(0xdeadbeef).String() = %q, want hex", got)
	}
}
