package rpcinterface_123456781234abcdef0001234567cffb_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "12345678-1234-abcd-ef00-01234567cffb" {
		t.Fatalf("UUID = %s, want 12345678-1234-abcd-ef00-01234567cffb", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the implemented opnums and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumNetrServerReqChallenge != 4 || OpnumNetrServerAuthenticate2 != 15 || OpnumNetrServerPasswordSet2 != 30 {
		t.Fatalf("opnums = %d/%d/%d, want 4/15/30", OpnumNetrServerReqChallenge, OpnumNetrServerAuthenticate2, OpnumNetrServerPasswordSet2)
	}
	if OpnumToName[15] != "NetrServerAuthenticate2" || NameToOpnum["NetrServerPasswordSet2"] != 30 {
		t.Fatal("opnum name mapping is inconsistent")
	}
}

// TestStatusCodesResolveThroughNTStatus verifies that the NTSTATUS values Netlogon
// documents resolve through the shared [MS-ERREF] 2.3.1 table under their NT_STATUS_*
// names, that a status the interface never enumerated now renders by name instead of as
// undecoded hex, and that a value no specification defines still renders as hex.
func TestStatusCodesResolveThroughNTStatus(t *testing.T) {
	documented := map[nt_status.NT_STATUS]string{
		0x00000000: "NT_STATUS_SUCCESS",
		0xC000000D: "NT_STATUS_INVALID_PARAMETER",
		0xC0000022: "NT_STATUS_ACCESS_DENIED",
		0xC0000064: "NT_STATUS_NO_SUCH_USER",
		0xC00000BB: "NT_STATUS_NOT_SUPPORTED",
		0xC000018B: "NT_STATUS_NO_TRUST_SAM_ACCOUNT",
		0x00000105: "NT_STATUS_MORE_ENTRIES",
		0x8000001A: "NT_STATUS_NO_MORE_ENTRIES",
		0xC0000388: "NT_STATUS_DOWNGRADE_DETECTED",
	}
	for status, want := range documented {
		if got := status.String(); got != want {
			t.Fatalf("NT_STATUS(0x%08x).String() = %s, want %s", uint32(status), got, want)
		}
	}
	// STATUS_NOLOGON_WORKSTATION_TRUST_ACCOUNT is a routine Netlogon rejection that the
	// interface's own subset never listed, so it used to print as bare hex.
	if got := nt_status.NT_STATUS(0xC0000199).String(); got != "NT_STATUS_NOLOGON_WORKSTATION_TRUST_ACCOUNT" {
		t.Fatalf("NT_STATUS(0xc0000199).String() = %s, want NT_STATUS_NOLOGON_WORKSTATION_TRUST_ACCOUNT", got)
	}
	if got := nt_status.NT_STATUS(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Fatalf("NT_STATUS(0xdeadbeef).String() = %s, want 0xdeadbeef", got)
	}
}
