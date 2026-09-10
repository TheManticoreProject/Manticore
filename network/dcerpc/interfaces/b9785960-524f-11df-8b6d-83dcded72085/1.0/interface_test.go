package rpcinterface_b9785960524f11df8b6d83dcded72085_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestStatusResolvesThroughHRESULTNotWin32 replaces the test that pinned this interface's
// status out of the Win32 migration. The table it was waiting for now exists, so the
// status resolves through windows/errors/hresult — and the reason it must still never be
// resolved through windows/errors/win32 is unchanged, so both halves are asserted here.
//
// GetKey is declared HRESULT in the [MS-GKDI] IDL and section 3.1.4.1 documents zero for
// success and a nonzero value for failure. HRESULTs are [MS-ERREF] section 2.1 and the
// WIN32_ERROR table is [MS-ERREF] section 2.2, and the two spaces disagree by value at
// both ends of the range.
func TestStatusResolvesThroughHRESULTNotWin32(t *testing.T) {
	// The success value [MS-GKDI] 3.1.4.1 documents resolves by name through the HRESULT
	// table, which is what the removed local constant and StatusString stood in for.
	if got := hresult.S_OK.String(); got != "S_OK" {
		t.Errorf("hresult.S_OK.String() = %q, want S_OK", got)
	}
	if uint32(hresult.S_OK) != 0x00000000 {
		t.Errorf("hresult.S_OK = 0x%08x, want 0x00000000 ([MS-GKDI] 3.1.4.1)", uint32(hresult.S_OK))
	}
	if err := hresult.S_OK.Error(); err != nil {
		t.Errorf("hresult.S_OK.Error() = %v, want nil", err)
	}

	// Failures GetKey can actually return now carry names instead of bare hex: the
	// FACILITY_WIN32 wrappings the shared table derives from the Win32 code in their low
	// half, and the cryptographic facility a key-service failure comes from.
	failures := map[hresult.HRESULT]string{
		0x80070005: "E_ACCESSDENIED",
		0x8007000D: "HRESULT_FROM_WIN32(ERROR_INVALID_DATA)",
		0x80090005: "NTE_BAD_DATA",
	}
	for code, name := range failures {
		if got := code.String(); got != name {
			t.Errorf("hresult.HRESULT(0x%08x).String() = %q, want %q", uint32(code), got, name)
		}
		if code.IsSuccess() {
			t.Errorf("hresult.HRESULT(0x%08x).IsSuccess() = true, want false", uint32(code))
		}
	}

	// The Win32 table names neither HRESULT mnemonic, so it never had anything to migrate
	// this interface's status to.
	for _, name := range []string{"S_OK", "S_FALSE"} {
		if code, defined := win32.FromName(name); defined {
			t.Errorf("win32.FromName(%q) resolved to 0x%08x; %s is an HRESULT and the Win32 table must not claim it", name, uint32(code), name)
		}
	}

	// At the low end the two spaces collide by value: an HRESULT with the severity bit
	// clear is a success code, S_FALSE being 0x00000001, and [MS-ERREF] 2.2 gives that
	// value an unrelated failure name. Reporting this interface's status out of that table
	// would read a success as an invalid-function failure.
	if entry, defined := win32.Lookup(win32.WIN32_ERROR(0x00000001)); !defined {
		t.Error("win32 defines no name for 0x00000001; [MS-ERREF] 2.2 gives it ERROR_INVALID_FUNCTION")
	} else if entry.Name != "ERROR_INVALID_FUNCTION" {
		t.Errorf("win32 names 0x00000001 %q, want ERROR_INVALID_FUNCTION: the collision this interface must keep out of its status reporting", entry.Name)
	}
	if got := hresult.S_FALSE.String(); got != "S_FALSE" {
		t.Errorf("hresult.S_FALSE.String() = %q, want S_FALSE: the HRESULT table is the one that reads 0x00000001 correctly", got)
	}
	if !hresult.S_FALSE.IsSuccess() {
		t.Error("hresult.S_FALSE.IsSuccess() = false, want true: success is a severity bit, not zero alone")
	}

	// At the high end the Win32 table simply does not reach, which is the other half of
	// why it could name no GKDI failure.
	for _, code := range []uint32{0x80070005, 0x8007000D, 0x80090005} {
		if _, defined := win32.Lookup(win32.WIN32_ERROR(code)); defined {
			t.Errorf("win32 names 0x%08x; the [MS-ERREF] 2.2 table must not claim an HRESULT failure", code)
		}
	}

	// A value neither table defines still renders as hex, and 0x00000002 — which the Win32
	// table calls ERROR_FILE_NOT_FOUND — is a success-severity HRESULT here, not a failure.
	if got := hresult.HRESULT(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("hresult.HRESULT(0xdeadbeef).String() = %q, want hex", got)
	}
	if got := hresult.HRESULT(0x00000002).String(); got != "0x00000002" {
		t.Errorf("hresult.HRESULT(0x00000002).String() = %q, want hex: the HRESULT space assigns it no name", got)
	}
	if !hresult.HRESULT(0x00000002).IsSuccess() {
		t.Error("hresult.HRESULT(0x00000002).IsSuccess() = false, want true: severity bit clear is success")
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// MS-GKDI defines a single on-the-wire method, GetKey (opnum 0).
	if len(OpnumToName) != 1 {
		t.Errorf("OpnumToName has %d entries, want 1 on-the-wire method", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d (a duplicate name collapsed an entry)",
			len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumGetKey] != "GetKey" {
		t.Errorf("OpnumToName[0] = %q, want GetKey", OpnumToName[OpnumGetKey])
	}
	if NameToOpnum["GetKey"] != OpnumGetKey || OpnumGetKey != 0 {
		t.Errorf("NameToOpnum[GetKey] = %d, want 0", NameToOpnum["GetKey"])
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// b9785960-524f-11df-8b6d-83dcded72085, version 1.0.
	if id.UUID.A != 0xb9785960 || id.UUID.B != 0x524f || id.UUID.C != 0x11df ||
		id.UUID.D != 0x8b6d || id.UUID.E != 0x83dcded72085 {
		t.Errorf("SyntaxID UUID = %+v, want b9785960-524f-11df-8b6d-83dcded72085", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestPipeName(t *testing.T) {
	if PipeName != `\lsass` {
		t.Errorf("PipeName = %q, want \\lsass", PipeName)
	}
}
