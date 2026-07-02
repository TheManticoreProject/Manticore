package rpcinterface_e3d0d746d2af40fd8a7a0d7078bb7092_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "e3d0d746-d2af-40fd-8a7a-0d7078bb7092" {
		t.Fatalf("UUID = %s, want e3d0d746-d2af-40fd-8a7a-0d7078bb7092", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the single opnum and its name mapping round-trips.
func TestOpnums(t *testing.T) {
	if OpnumExchangePublicKeys != 0 {
		t.Fatalf("OpnumExchangePublicKeys = %d, want 0", OpnumExchangePublicKeys)
	}
	if OpnumToName[0] != "ExchangePublicKeys" || NameToOpnum["ExchangePublicKeys"] != 0 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("map sizes differ: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering of the documented HRESULTs and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Fatalf("StatusString(Success) = %s, want ERROR_SUCCESS", got)
	}
	if got := StatusString(StatusAccessDenied); got != "E_ACCESSDENIED" {
		t.Fatalf("StatusString(AccessDenied) = %s, want E_ACCESSDENIED", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
