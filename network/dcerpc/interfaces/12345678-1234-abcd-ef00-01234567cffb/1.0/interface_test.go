package rpcinterface_123456781234abcdef0001234567cffb_1_0

import "testing"

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

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusAccessDenied); got != "STATUS_ACCESS_DENIED" {
		t.Fatalf("StatusString(AccessDenied) = %s", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
