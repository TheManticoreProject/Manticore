package rpcinterface_22e5386d8b124bf0b0ec6a1ea419e366_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "22e5386d-8b12-4bf0-b0ec-6a1ea419e366" {
		t.Fatalf("UUID = %s, want 22e5386d-8b12-4bf0-b0ec-6a1ea419e366", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the implemented opnums and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumRpcNetEventOpenSession != 0 || OpnumRpcNetEventReceiveData != 1 || OpnumRpcNetEventCloseSession != 2 {
		t.Fatalf("opnums = %d/%d/%d, want 0/1/2", OpnumRpcNetEventOpenSession, OpnumRpcNetEventReceiveData, OpnumRpcNetEventCloseSession)
	}
	if OpnumToName[0] != "RpcNetEventOpenSession" || NameToOpnum["RpcNetEventCloseSession"] != 2 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Fatalf("StatusString(StatusSuccess) = %s, want ERROR_SUCCESS", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
