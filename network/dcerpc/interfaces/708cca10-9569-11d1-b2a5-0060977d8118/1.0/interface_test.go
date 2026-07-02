package rpcinterface_708cca10956911d1b2a50060977d8118_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the dscomm2 interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "708cca10-9569-11d1-b2a5-0060977d8118" {
		t.Fatalf("UUID = %s, want 708cca10-9569-11d1-b2a5-0060977d8118", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the implemented opnums (opnum 7 is not used on the wire) and the
// name mapping.
func TestOpnums(t *testing.T) {
	if OpnumS_DSGetComputerSites != 0 || OpnumS_DSIsServerGC != 6 || OpnumS_DSGetGCListInDomain != 8 {
		t.Fatalf("opnums = %d/%d/%d, want 0/6/8", OpnumS_DSGetComputerSites, OpnumS_DSIsServerGC, OpnumS_DSGetGCListInDomain)
	}
	if OpnumToName[0] != "S_DSGetComputerSites" || NameToOpnum["S_DSGetGCListInDomain"] != 8 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if _, ok := OpnumToName[7]; ok {
		t.Fatal("opnum 7 is not used on the wire and must not be mapped")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 8 {
		t.Fatalf("on-the-wire method count = %d, want 8", len(OpnumToName))
	}
}

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "MQ_OK" {
		t.Fatalf("StatusString(StatusSuccess) = %s, want MQ_OK", got)
	}
	if got := StatusString(MQ_ERROR_NO_DS); got != "MQ_ERROR_NO_DS" {
		t.Fatalf("StatusString(MQ_ERROR_NO_DS) = %s, want MQ_ERROR_NO_DS", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
