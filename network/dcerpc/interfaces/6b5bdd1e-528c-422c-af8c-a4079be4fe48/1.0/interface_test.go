package rpcinterface_6b5bdd1e528c422caf8ca4079be4fe48_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID pins the abstract syntax identifier for the RemoteFW interface
// (6b5bdd1e-528c-422c-af8c-a4079be4fe48 v1.0, [MS-FASP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "6b5bdd1e-528c-422c-af8c-a4079be4fe48" {
		t.Errorf("UUID = %s, want 6b5bdd1e-528c-422c-af8c-a4079be4fe48", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// that the anchor opnums resolve to the expected method names. RemoteFW exposes 94
// contiguous opnums (0..93); none are "not used on the wire".
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) differ in size",
			len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 94 {
		t.Fatalf("OpnumToName has %d entries, want 94", len(OpnumToName))
	}
	if OpnumToName[OpnumRRPC_FWOpenPolicyStore] != "RRPC_FWOpenPolicyStore" {
		t.Errorf("opnum 0 = %q, want RRPC_FWOpenPolicyStore", OpnumToName[OpnumRRPC_FWOpenPolicyStore])
	}
	if OpnumToName[OpnumRRPC_FWQueryFirewallRules2_33] != "RRPC_FWQueryFirewallRules2_33" {
		t.Errorf("opnum 93 = %q, want RRPC_FWQueryFirewallRules2_33", OpnumToName[OpnumRRPC_FWQueryFirewallRules2_33])
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
}

// TestOpnumContiguous checks the opnum space is exactly 0..93 with no gaps.
func TestOpnumContiguous(t *testing.T) {
	for op := uint16(0); op <= 93; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	if _, ok := OpnumToName[94]; ok {
		t.Errorf("opnum 94 should not exist")
	}
}

// TestDocumentedStatusCodesResolveThroughWin32 checks that every Win32 error code
// [MS-FASP] documents this interface returning renders under its [MS-ERREF] 2.2
// symbolic name through the shared win32 table, which is where these codes now live
// instead of a private subset in this package.
func TestDocumentedStatusCodesResolveThroughWin32(t *testing.T) {
	documented := []struct {
		code uint32
		name string
	}{
		{0x00000000, "ERROR_SUCCESS"},
		{0x00000002, "ERROR_FILE_NOT_FOUND"},
		{0x00000005, "ERROR_ACCESS_DENIED"},
		{0x00000006, "ERROR_INVALID_HANDLE"},
		{0x00000008, "ERROR_NOT_ENOUGH_MEMORY"},
		{0x0000000D, "ERROR_INVALID_DATA"},
		{0x00000050, "ERROR_FILE_EXISTS"},
		{0x00000057, "ERROR_INVALID_PARAMETER"},
		{0x0000007A, "ERROR_INSUFFICIENT_BUFFER"},
		{0x000000B7, "ERROR_ALREADY_EXISTS"},
		{0x00000103, "ERROR_NO_MORE_ITEMS"},
		{0x00000490, "ERROR_NOT_FOUND"},
	}
	for _, tc := range documented {
		if got := win32.WIN32_ERROR(tc.code).String(); got != tc.name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", tc.code, got, tc.name)
		}
		if got, ok := win32.FromName(tc.name); !ok || uint32(got) != tc.code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", tc.name, uint32(got), ok, tc.code)
		}
	}
}

// TestStatusCodeOutsideOldSubsetRendersByName covers the reason the private subset
// was a ceiling: a code the interface can return but that the subset never listed
// used to render as a bare hex value, and now resolves to its specification name.
// ERROR_INVALID_LEVEL (0x0000007C) and ERROR_SERVICE_DOES_NOT_EXIST (0x00000424)
// were both outside the twelve-code subset.
func TestStatusCodeOutsideOldSubsetRendersByName(t *testing.T) {
	if got := win32.WIN32_ERROR(0x0000007C).String(); got != "ERROR_INVALID_LEVEL" {
		t.Errorf("win32.WIN32_ERROR(0x0000007c).String() = %q, want ERROR_INVALID_LEVEL", got)
	}
	if got := win32.WIN32_ERROR(0x00000424).String(); got != "ERROR_SERVICE_DOES_NOT_EXIST" {
		t.Errorf("win32.WIN32_ERROR(0x00000424).String() = %q, want ERROR_SERVICE_DOES_NOT_EXIST", got)
	}
}

// TestUndefinedStatusCodeFallsBackToHex checks the fallback for a value [MS-ERREF]
// 2.2 defines no name for: the highest code in the table is 0x00003BC3, so
// 0xDEADBEEF is undefined and must render as hex rather than as an empty string.
func TestUndefinedStatusCodeFallsBackToHex(t *testing.T) {
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want 0xdeadbeef", got)
	}
	if name := win32.WIN32_ERROR(0xDEADBEEF).Name(); name != "" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).Name() = %q, want empty", name)
	}
}

// TestPipeName documents that this interface has no named pipe: MS-FASP uses
// ncacn_ip_tcp with a dynamic endpoint, so PipeName is intentionally empty.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}
