package rpcinterface_4b324fc8167001d312785a47bf6ee188_3_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestDocumentedStatusCodesResolveThroughWin32 pins the NET_API_STATUS codes [MS-SRVS]
// documents to the shared [MS-ERREF] 2.2 table: each resolves under the symbolic name
// the specification gives it, and each renders under a name rather than a number.
func TestDocumentedStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[string]win32.WIN32_ERROR{
		"NERR_Success":            0,
		"ERROR_SUCCESS":           0,
		"ERROR_FILE_NOT_FOUND":    2,
		"ERROR_ACCESS_DENIED":     5,
		"ERROR_NOT_SUPPORTED":     50,
		"ERROR_INVALID_PARAMETER": 87,
		"ERROR_INVALID_NAME":      123,
		"ERROR_INVALID_LEVEL":     124,
		"ERROR_MORE_DATA":         234,
		"NERR_BufTooSmall":        2123,
		"NERR_DuplicateShare":     2118,
		"NERR_UserNotFound":       2221,
		"NERR_NetNameNotFound":    2310,
		"NERR_DeviceNotShared":    2311,
		"NERR_ClientNameNotFound": 2312,
	}
	for name, code := range documented {
		got, defined := win32.FromName(name)
		if !defined {
			t.Errorf("win32.FromName(%q) is undefined", name)
			continue
		}
		if got != code {
			t.Errorf("win32.FromName(%q) = %d, want %d", name, got, code)
		}
		if rendered := code.String(); rendered == "" || rendered[0] == '0' {
			t.Errorf("win32.WIN32_ERROR(%d).String() = %q, want a symbolic name", code, rendered)
		}
	}
}

// TestStatusCodeOutsideOldSubsetRendersByName covers what the interface's private
// subset could not: a Win32 error srvsvc can return that the subset never listed now
// decodes to its name instead of an undecoded number, while a value [MS-ERREF] 2.2
// leaves undefined still renders as hexadecimal.
func TestStatusCodeOutsideOldSubsetRendersByName(t *testing.T) {
	// ERROR_NETNAME_DELETED (64) and NERR_ServerNotStarted (2114) were outside the
	// interface's old subset and used to print as "64" and "2114".
	if got := win32.WIN32_ERROR(64).String(); got != "ERROR_NETNAME_DELETED" {
		t.Errorf("win32.WIN32_ERROR(64).String() = %q, want ERROR_NETNAME_DELETED", got)
	}
	if got := win32.WIN32_ERROR(2114).String(); got != "NERR_ServerNotStarted" {
		t.Errorf("win32.WIN32_ERROR(2114).String() = %q, want NERR_ServerNotStarted", got)
	}
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hexadecimal", got)
	}
}

// TestNerrBaseStaysLocal guards the one constant that could not migrate: NERR_BASE is
// the base of the lmerr.h network-error range, not a code, so [MS-ERREF] 2.2 has no row
// for it.
func TestNerrBaseStaysLocal(t *testing.T) {
	if NERR_BASE != 2100 {
		t.Errorf("NERR_BASE = %d, want 2100", NERR_BASE)
	}
	if _, defined := win32.Lookup(win32.WIN32_ERROR(NERR_BASE)); defined {
		t.Error("win32 defines a code for NERR_BASE; it no longer needs a local declaration")
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0x4b324fc8 || id.UUID.B != 0x1670 || id.UUID.C != 0x01d3 ||
		id.UUID.D != 0x1278 || id.UUID.E != 0x5a47bf6ee188 {
		t.Errorf("SyntaxID UUID = %+v, want 4b324fc8-1670-01d3-1278-5a47bf6ee188", id.UUID)
	}
	if id.MajorVersion != 3 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 3.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 47 {
		t.Errorf("OpnumToName has %d entries, want 47 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumNetrShareEnum] != "NetrShareEnum" {
		t.Errorf("OpnumToName[15] = %q", OpnumToName[OpnumNetrShareEnum])
	}
}
