package rpcinterface_afc07e2e311c4435808cc483ffeec7c9_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// TestSyntaxID pins the abstract syntax identifier for the lsacap interface
// (afc07e2e-311c-4435-808c-c483ffeec7c9 v1.0, [MS-CAPR]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "afc07e2e-311c-4435-808c-c483ffeec7c9" {
		t.Errorf("UUID = %s, want afc07e2e-311c-4435-808c-c483ffeec7c9", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that the single on-the-wire opnum (0) is covered.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 1 {
		t.Fatalf("OpnumToName has %d entries, want 1 (opnum 0)", len(OpnumToName))
	}
	if OpnumToName[OpnumLsarGetAvailableCAPIDs] != "LsarGetAvailableCAPIDs" {
		t.Errorf("opnum 0 = %q, want LsarGetAvailableCAPIDs", OpnumToName[OpnumLsarGetAvailableCAPIDs])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestPipeName checks lsacap rides the shared LSA pipe.
func TestPipeName(t *testing.T) {
	if PipeName != `\lsarpc` {
		t.Errorf("PipeName = %q, want \\lsarpc", PipeName)
	}
}

// TestStatusCodesResolveThroughSharedTable checks that the status codes
// [MS-CAPR] 3.1.4.1 documents for this interface resolve through the shared
// [MS-ERREF] 2.3.1 table under their NT_STATUS_* names, that a status outside
// the subset this package used to declare now resolves by name instead of
// rendering as undecoded hex, and that a value the specification does not
// define still renders as hex.
func TestStatusCodesResolveThroughSharedTable(t *testing.T) {
	documented := []struct {
		status nt_status.NT_STATUS
		name   string
	}{
		{nt_status.NT_STATUS_SUCCESS, "NT_STATUS_SUCCESS"},
		{nt_status.NT_STATUS_ACCESS_DENIED, "NT_STATUS_ACCESS_DENIED"},
		{nt_status.NT_STATUS_INVALID_PARAMETER, "NT_STATUS_INVALID_PARAMETER"},
		{nt_status.NT_STATUS_INSUFFICIENT_RESOURCES, "NT_STATUS_INSUFFICIENT_RESOURCES"},
	}
	for _, tc := range documented {
		if got := tc.status.String(); got != tc.name {
			t.Errorf("NT_STATUS(0x%08x).String() = %q, want %q", uint32(tc.status), got, tc.name)
		}
	}

	// Outside the four codes the old private subset carried, so the previous
	// StatusString rendered it as "0xc0000034".
	if got := nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND.String(); got != "NT_STATUS_OBJECT_NAME_NOT_FOUND" {
		t.Errorf("NT_STATUS(0xc0000034).String() = %q, want NT_STATUS_OBJECT_NAME_NOT_FOUND", got)
	}

	if got := nt_status.NT_STATUS(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("NT_STATUS(0xdeadbeef).String() = %q, want 0xdeadbeef", got)
	}
}
