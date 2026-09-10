package rpcinterface_4fc742e04a1011cf827300aa004ae673_3_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestDocumentedStatusCodesResolveThroughWin32 pins the NET_API_STATUS codes [MS-DFSNM]
// documents to the shared [MS-ERREF] 2.2 table: each resolves under the symbolic name
// the specification gives it, and each renders under a name rather than a number.
func TestDocumentedStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[string]win32.WIN32_ERROR{
		"NERR_Success":            0,
		"ERROR_SUCCESS":           0,
		"ERROR_FILE_NOT_FOUND":    2,
		"ERROR_ACCESS_DENIED":     5,
		"ERROR_NOT_ENOUGH_MEMORY": 8,
		"ERROR_NOT_SUPPORTED":     50,
		"ERROR_FILE_EXISTS":       80,
		"ERROR_INVALID_PARAMETER": 87,
		"ERROR_INVALID_NAME":      123,
		"ERROR_DIR_NOT_EMPTY":     145,
		"ERROR_ALREADY_EXISTS":    183,
		"ERROR_NOT_FOUND":         1168,
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
// subset could not: a Win32 error netdfs can return that the subset never listed now
// decodes to its name instead of an undecoded number, while a value [MS-ERREF] 2.2
// leaves undefined still renders as hexadecimal.
func TestStatusCodeOutsideOldSubsetRendersByName(t *testing.T) {
	// NERR_DfsNoSuchVolume (2662) and ERROR_INVALID_LEVEL (124) were outside the
	// interface's old subset and used to print as "0x00000a66" and "0x0000007c".
	if got := win32.WIN32_ERROR(2662).String(); got != "NERR_DfsNoSuchVolume" {
		t.Errorf("win32.WIN32_ERROR(2662).String() = %q, want NERR_DfsNoSuchVolume", got)
	}
	if got := win32.WIN32_ERROR(124).String(); got != "ERROR_INVALID_LEVEL" {
		t.Errorf("win32.WIN32_ERROR(124).String() = %q, want ERROR_INVALID_LEVEL", got)
	}
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hexadecimal", got)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// MS-DFSNM defines 26 opnums, of which 3 (7, 8, 9) are "not used on the wire".
	if len(OpnumToName) != 23 {
		t.Errorf("OpnumToName has %d entries, want 23 on-the-wire methods", len(OpnumToName))
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
	// The 7/8/9 gap must not have leaked a mapping.
	for _, gap := range []uint16{7, 8, 9} {
		if name, ok := OpnumToName[gap]; ok {
			t.Errorf("OpnumToName[%d] = %q, want absent (not used on the wire)", gap, name)
		}
	}
	// Spot-check both directions, including a post-gap opnum.
	if OpnumToName[OpnumNetrDfsManagerGetVersion] != "NetrDfsManagerGetVersion" {
		t.Errorf("OpnumToName[0] = %q, want NetrDfsManagerGetVersion", OpnumToName[OpnumNetrDfsManagerGetVersion])
	}
	if NameToOpnum["NetrDfsAddFtRoot"] != OpnumNetrDfsAddFtRoot || OpnumNetrDfsAddFtRoot != 10 {
		t.Errorf("NameToOpnum[NetrDfsAddFtRoot] = %d, want 10", NameToOpnum["NetrDfsAddFtRoot"])
	}
	if NameToOpnum["NetrDfsGetSupportedNamespaceVersion"] != 25 {
		t.Errorf("NameToOpnum[NetrDfsGetSupportedNamespaceVersion] = %d, want 25", NameToOpnum["NetrDfsGetSupportedNamespaceVersion"])
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// 4fc742e0-4a10-11cf-8273-00aa004ae673, version 3.0.
	if id.UUID.A != 0x4fc742e0 || id.UUID.B != 0x4a10 || id.UUID.C != 0x11cf ||
		id.UUID.D != 0x8273 || id.UUID.E != 0x00aa004ae673 {
		t.Errorf("SyntaxID UUID = %+v, want 4fc742e0-4a10-11cf-8273-00aa004ae673", id.UUID)
	}
	if id.MajorVersion != 3 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 3.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestPipeName(t *testing.T) {
	if PipeName != `\netdfs` {
		t.Errorf("PipeName = %q, want \\netdfs", PipeName)
	}
}
