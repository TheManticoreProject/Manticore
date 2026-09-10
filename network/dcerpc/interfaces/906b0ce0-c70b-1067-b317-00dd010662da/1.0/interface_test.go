package rpcinterface_906b0ce0c70b1067b31700dd010662da_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "906b0ce0-c70b-1067-b317-00dd010662da" {
		t.Fatalf("UUID = %s, want 906b0ce0-c70b-1067-b317-00dd010662da", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the eight opnums are contiguous 0..7 and the name maps round-trip.
func TestOpnums(t *testing.T) {
	want := map[uint16]string{
		0: "Poke",
		1: "BuildContext",
		2: "NegotiateResources",
		3: "SendReceive",
		4: "TearDownContext",
		5: "BeginTearDown",
		6: "PokeW",
		7: "BuildContextW",
	}
	if len(OpnumToName) != len(want) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(want))
	}
	for op, name := range want {
		if OpnumToName[op] != name {
			t.Errorf("OpnumToName[%d] = %q, want %q", op, OpnumToName[op], name)
		}
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("map sizes differ: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughHRESULT pins that the four HRESULTs this interface used
// to declare resolve through the shared [MS-ERREF] 2.1.1 table under the names the
// specification gives them, that the table names values the old four-value subset could
// only render as hex, and that a value the specification does not define still renders
// as hex.
func TestStatusCodesResolveThroughHRESULT(t *testing.T) {
	documented := map[hresult.HRESULT]string{
		0x00000000: "S_OK",
		0x80004005: "E_FAIL",
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

	// E_ACCESSDENIED is one a [MS-CMPO] partner may surface and the old subset did not
	// declare, so it rendered as hex; it resolves by name now.
	if got := hresult.HRESULT(0x80070005).String(); got != "E_ACCESSDENIED" {
		t.Errorf("hresult.HRESULT(0x80070005).String() = %q, want E_ACCESSDENIED", got)
	}

	// A value [MS-ERREF] 2.1.1 does not define, and whose facility is not FACILITY_WIN32
	// either, still renders as hex.
	if got := hresult.HRESULT(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("hresult.HRESULT(0xdeadbeef).String() = %q, want hex", got)
	}

	// The stubs compare against S_OK rather than testing the severity bit, and that
	// distinction is real: S_FALSE is a success value too, so IsSuccess accepts a status
	// the comparison rejects.
	if !hresult.S_FALSE.IsSuccess() {
		t.Error("hresult.S_FALSE.IsSuccess() = false, want true")
	}
	if hresult.S_FALSE == hresult.S_OK {
		t.Error("hresult.S_FALSE == hresult.S_OK, want distinct values")
	}
}
