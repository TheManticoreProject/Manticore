package rpcinterface_1a9134dd7b3945baad8844d01ca47f28_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the RemoteRead interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "1a9134dd-7b39-45ba-ad88-44d01ca47f28" {
		t.Fatalf("UUID = %s, want 1a9134dd-7b39-45ba-ad88-44d01ca47f28", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the (empty) transport endpoint: RemoteRead is ncacn_ip_tcp only.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}

// TestOpnums verifies the opnums and the name mapping. Opnum 1 (Opnum1NotUsedOnWire) is
// absent by design.
func TestOpnums(t *testing.T) {
	if OpnumR_GetServerPort != 0 || OpnumR_OpenQueue != 2 || OpnumR_EndTransactionalReceive != 15 {
		t.Fatalf("opnums = %d/%d/%d, want 0/2/15", OpnumR_GetServerPort, OpnumR_OpenQueue, OpnumR_EndTransactionalReceive)
	}
	if OpnumToName[0] != "R_GetServerPort" || NameToOpnum["R_EndTransactionalReceive"] != 15 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if _, ok := OpnumToName[1]; ok {
		t.Fatal("opnum 1 must be absent (Opnum1NotUsedOnWire)")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 15 {
		t.Fatalf("on-the-wire method count = %d, want 15", len(OpnumToName))
	}
}

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "MQ_OK" {
		t.Fatalf("StatusString(StatusSuccess) = %s, want MQ_OK", got)
	}
	if got := StatusString(MQ_ERROR_INVALID_HANDLE); got != "MQ_ERROR_INVALID_HANDLE" {
		t.Fatalf("StatusString(MQ_ERROR_INVALID_HANDLE) = %s, want MQ_ERROR_INVALID_HANDLE", got)
	}
	if got := StatusString(MQ_ERROR_MESSAGE_ALREADY_RECEIVED); got != "MQ_ERROR_MESSAGE_ALREADY_RECEIVED" {
		t.Fatalf("StatusString(MQ_ERROR_MESSAGE_ALREADY_RECEIVED) = %s, want MQ_ERROR_MESSAGE_ALREADY_RECEIVED", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
