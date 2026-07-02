package rpcinterface_3dde7c30165d11d1ab8f00805f14db40_1_0

import "testing"

func TestStatusString(t *testing.T) {
	if got := StatusString(ErrorSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(ErrorInvalidParameter); got != "ERROR_INVALID_PARAMETER" {
		t.Errorf("StatusString(0x57) = %q, want ERROR_INVALID_PARAMETER", got)
	}
	if got := StatusString(0x12345678); got != "0x12345678" {
		t.Errorf("StatusString(unknown) = %q, want hex fallback", got)
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
