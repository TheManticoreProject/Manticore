package rpcinterface_338cd001224431f1aaaa900038001003_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestDocumentedStatusCodesResolveThroughWin32 pins the Win32 codes [MS-RRP] 3.1.5
// documents for the winreg methods to the names the shared [MS-ERREF] 2.2 table gives
// them. The interface no longer declares a subset of its own; the codes are looked up by
// value in win32, so a code the interface never listed still renders by name and only a
// value the specification leaves undefined falls back to hex.
func TestDocumentedStatusCodesResolveThroughWin32(t *testing.T) {
	documented := []struct {
		code win32.WIN32_ERROR
		name string
	}{
		{0x00000000, "ERROR_SUCCESS"},
		{0x00000002, "ERROR_FILE_NOT_FOUND"},
		{0x00000005, "ERROR_ACCESS_DENIED"},
		{0x00000006, "ERROR_INVALID_HANDLE"},
		{0x00000057, "ERROR_INVALID_PARAMETER"},
		{0x00000078, "ERROR_CALL_NOT_IMPLEMENTED"},
		{0x0000007A, "ERROR_INSUFFICIENT_BUFFER"},
		{0x000000A1, "ERROR_BAD_PATHNAME"},
		{0x000000EA, "ERROR_MORE_DATA"},
		{0x00000103, "ERROR_NO_MORE_ITEMS"},
		{0x000003F9, "ERROR_NOT_REGISTRY_FILE"},
		{0x000003FA, "ERROR_KEY_DELETED"},
		{0x000003FB, "ERROR_NO_LOG_SPACE"},
		{0x0000045B, "ERROR_SHUTDOWN_IN_PROGRESS"},
	}
	for _, c := range documented {
		if got := c.code.String(); got != c.name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", uint32(c.code), got, c.name)
		}
	}

	// Outside the subset the interface used to carry: ERROR_KEY_HAS_CHILDREN (0x000003FC)
	// and ERROR_BADKEY (0x000003F2) were rendered as undecoded hex before and now resolve
	// by name.
	for _, c := range []struct {
		code win32.WIN32_ERROR
		name string
	}{
		{0x000003FC, "ERROR_KEY_HAS_CHILDREN"},
		{0x000003F2, "ERROR_BADKEY"},
	} {
		if got := c.code.String(); got != c.name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", uint32(c.code), got, c.name)
		}
	}

	// A value the specification defines no name for still renders, as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0x338cd001 || id.UUID.B != 0x2244 || id.UUID.C != 0x31f1 ||
		id.UUID.D != 0xaaaa || id.UUID.E != 0x900038001003 {
		t.Errorf("SyntaxID UUID = %+v, want 338cd001-2244-31f1-aaaa-900038001003", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// 36 opnum slots (0-35), including the 5 reserved NotImplemented slots kept for
	// wire-numbering completeness.
	if len(OpnumToName) != 36 {
		t.Errorf("OpnumToName has %d entries, want 36 opnum slots", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumBaseRegOpenKey] != "BaseRegOpenKey" || OpnumBaseRegOpenKey != 15 {
		t.Errorf("OpnumBaseRegOpenKey = %d (%q)", OpnumBaseRegOpenKey, OpnumToName[OpnumBaseRegOpenKey])
	}
	if OpnumOpenLocalMachine != 2 || OpnumBaseRegQueryValue != 17 || OpnumBaseRegDeleteKeyEx != 35 {
		t.Errorf("opnum numbering off: OpenLocalMachine=%d QueryValue=%d DeleteKeyEx=%d",
			OpnumOpenLocalMachine, OpnumBaseRegQueryValue, OpnumBaseRegDeleteKeyEx)
	}
}
