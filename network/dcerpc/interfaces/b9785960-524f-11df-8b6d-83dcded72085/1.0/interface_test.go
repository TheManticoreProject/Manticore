package rpcinterface_b9785960524f11df8b6d83dcded72085_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "S_OK" {
		t.Errorf("StatusString(0) = %q, want S_OK", got)
	}
	if got := StatusString(0x80070005); got != "0x80070005" {
		t.Errorf("StatusString(unknown) = %q, want hex fallback", got)
	}
}

// TestStatusIsHRESULTNotWin32 pins this interface's status to the HRESULT space it comes
// from, so that a later pass does not route it through windows/errors/win32. GetKey is
// declared HRESULT in the IDL and [MS-GKDI] section 3.1.4.1 documents zero for success and
// a nonzero value for failure; HRESULTs are [MS-ERREF] section 2.1 and the WIN32_ERROR
// table is [MS-ERREF] section 2.2, and the two spaces disagree at both ends of the range.
func TestStatusIsHRESULTNotWin32(t *testing.T) {
	if StatusSuccess != 0x00000000 {
		t.Errorf("StatusSuccess = 0x%08x, want 0x00000000 (S_OK, [MS-GKDI] 3.1.4.1)", StatusSuccess)
	}

	// The Win32 table names neither HRESULT mnemonic, so a migration would have nothing
	// to migrate this interface's status to.
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

	// At the high end the table simply does not reach. A failing GetKey returns an HRESULT
	// with the severity bit set, typically an HRESULT_FROM_WIN32 wrapping in the 0x8007xxxx
	// range ([MS-ERREF] 2.1.2), and the shared table names none of those values.
	for _, code := range []uint32{0x80070005, 0x8007000D, 0x80090005} {
		if _, defined := win32.Lookup(win32.WIN32_ERROR(code)); defined {
			t.Errorf("win32 names 0x%08x; the [MS-ERREF] 2.2 table must not claim an HRESULT failure", code)
		}
	}

	// StatusString renders only the HRESULT name and never falls through to the Win32
	// table: 0x00000002 is ERROR_FILE_NOT_FOUND there and has no meaning here.
	if got := StatusString(0x00000002); got != "0x00000002" {
		t.Errorf("StatusString(0x00000002) = %q, want 0x00000002: the HRESULT return must not be named out of the Win32 table", got)
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
