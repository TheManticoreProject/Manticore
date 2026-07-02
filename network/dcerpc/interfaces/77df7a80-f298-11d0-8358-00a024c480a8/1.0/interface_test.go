package rpcinterface_77df7a80f29811d0835800a024c480a8_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the dscomm interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "77df7a80-f298-11d0-8358-00a024c480a8" {
		t.Fatalf("UUID = %s, want 77df7a80-f298-11d0-8358-00a024c480a8", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies representative opnums (including the gaps left by the "not used on
// the wire" and [callback] methods) and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumS_DSCreateObject != 0 || OpnumS_DSLookupEnd != 8 || OpnumS_DSDeleteObjectGuid != 10 {
		t.Fatalf("opnums = %d/%d/%d, want 0/8/10", OpnumS_DSCreateObject, OpnumS_DSLookupEnd, OpnumS_DSDeleteObjectGuid)
	}
	if OpnumS_DSQMSetMachineProperties != 19 || OpnumS_DSValidateServer != 22 || OpnumS_DSGetServerPort != 27 {
		t.Fatalf("opnums = %d/%d/%d, want 19/22/27", OpnumS_DSQMSetMachineProperties, OpnumS_DSValidateServer, OpnumS_DSGetServerPort)
	}
	if OpnumToName[0] != "S_DSCreateObject" || NameToOpnum["S_DSGetServerPort"] != 27 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 20 {
		t.Fatalf("on-the-wire method count = %d, want 20", len(OpnumToName))
	}
}

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "MQ_OK" {
		t.Fatalf("StatusString(StatusSuccess) = %s, want MQ_OK", got)
	}
	if got := StatusString(MQ_ERROR_QUEUE_NOT_FOUND); got != "MQ_ERROR_QUEUE_NOT_FOUND" {
		t.Fatalf("StatusString(MQ_ERROR_QUEUE_NOT_FOUND) = %s, want MQ_ERROR_QUEUE_NOT_FOUND", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
