package rpcinterface_3dde7c30165d11d1ab8f00805f14db40_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestStatusCodesResolveThroughWin32 pins that the two NET_API_STATUS codes [MS-BKRP]
// section 3.1.4.1 names for BackuprKey resolve through the shared [MS-ERREF] 2.2 table
// under their specification names, that codes the interface never enumerated now render
// by name rather than as undecoded hex, and that a value the specification does not
// define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000057: "ERROR_INVALID_PARAMETER",
	}
	if len(documented) != 2 {
		t.Fatalf("documented table has %d entries, want the 2 codes the descriptor declared", len(documented))
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// The descriptor called every other nonzero code opaque, so a server refusing the
	// call or rejecting the input BLOB rendered as undecoded hex. These resolve by name
	// now, which is the whole point of routing through the shared table.
	outsideOldSubset := map[uint32]string{
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x0000054F: "ERROR_INTERNAL_ERROR",
	}
	for code, name := range outsideOldSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0x12345678).String(); got != "0x12345678" {
		t.Errorf("win32.WIN32_ERROR(0x12345678).String() = %q, want hex", got)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	// MS-BKRP defines a single on-the-wire method, BackuprKey (opnum 0).
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
	if OpnumToName[OpnumBackuprKey] != "BackuprKey" {
		t.Errorf("OpnumToName[0] = %q, want BackuprKey", OpnumToName[OpnumBackuprKey])
	}
	if NameToOpnum["BackuprKey"] != OpnumBackuprKey || OpnumBackuprKey != 0 {
		t.Errorf("NameToOpnum[BackuprKey] = %d, want 0", NameToOpnum["BackuprKey"])
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// 3dde7c30-165d-11d1-ab8f-00805f14db40, version 1.0.
	if id.UUID.A != 0x3dde7c30 || id.UUID.B != 0x165d || id.UUID.C != 0x11d1 ||
		id.UUID.D != 0xab8f || id.UUID.E != 0x00805f14db40 {
		t.Errorf("SyntaxID UUID = %+v, want 3dde7c30-165d-11d1-ab8f-00805f14db40", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestActionAgentGUIDs(t *testing.T) {
	// Well-known action-agent GUIDs from [MS-BKRP] section 3.1.4.1.
	if BackupKeyBackupGUID.ToFormatD() != "7f752b10-178e-11d1-ab8f-00805f14db40" {
		t.Errorf("BackupKeyBackupGUID = %s, want 7f752b10-178e-11d1-ab8f-00805f14db40", BackupKeyBackupGUID.ToFormatD())
	}
	if BackupKeyRestoreGUIDWin2K.ToFormatD() != "7fe94d50-178e-11d1-ab8f-00805f14db40" {
		t.Errorf("BackupKeyRestoreGUIDWin2K = %s, want 7fe94d50-178e-11d1-ab8f-00805f14db40", BackupKeyRestoreGUIDWin2K.ToFormatD())
	}
	if BackupKeyRetrieveBackupKeyGUID.ToFormatD() != "018ff48a-eaba-40c6-8f6d-72370240e967" {
		t.Errorf("BackupKeyRetrieveBackupKeyGUID = %s, want 018ff48a-eaba-40c6-8f6d-72370240e967", BackupKeyRetrieveBackupKeyGUID.ToFormatD())
	}
	if BackupKeyRestoreGUID.ToFormatD() != "47270c64-2fc7-499b-ac5b-0e37cdce899a" {
		t.Errorf("BackupKeyRestoreGUID = %s, want 47270c64-2fc7-499b-ac5b-0e37cdce899a", BackupKeyRestoreGUID.ToFormatD())
	}
}
