package rpcinterface_c681d488d85011d08c5200c04fd90f7e_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestStatusCodesResolveThroughWin32 pins that the Win32 codes [MS-EFSR] 3.1.4
// documents for efsrpc resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a code the interface never enumerated now renders by name
// rather than as undecoded hex, and that a value the specification does not define
// still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x000000EA: "ERROR_MORE_DATA",
		0x00000103: "ERROR_NO_MORE_ITEMS",
		0x00001770: "ERROR_ENCRYPTION_FAILED",
		0x00001771: "ERROR_DECRYPTION_FAILED",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// ERROR_FILE_ENCRYPTED was outside the subset this interface used to declare, so it
	// rendered as hex; it resolves by name now.
	if got := win32.WIN32_ERROR(0x00001772).String(); got != "ERROR_FILE_ENCRYPTED" {
		t.Errorf("win32.WIN32_ERROR(0x00001772).String() = %q, want ERROR_FILE_ENCRYPTED", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0xc681d488 || id.UUID.B != 0xd850 || id.UUID.C != 0x11d0 ||
		id.UUID.D != 0x8c52 || id.UUID.E != 0x00c04fd90f7e {
		t.Errorf("SyntaxID UUID = %+v, want c681d488-d850-11d0-8c52-00c04fd90f7e", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 20 {
		t.Errorf("OpnumToName has %d entries, want 20 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumEfsRpcOpenFileRaw] != "EfsRpcOpenFileRaw" {
		t.Errorf("OpnumToName[0] = %q", OpnumToName[OpnumEfsRpcOpenFileRaw])
	}
}
