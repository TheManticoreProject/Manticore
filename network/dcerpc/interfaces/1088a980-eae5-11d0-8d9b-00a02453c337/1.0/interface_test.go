package rpcinterface_1088a980eae511d08d9b00a02453c337_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the qm2qm interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "1088a980-eae5-11d0-8d9b-00a02453c337" {
		t.Fatalf("UUID = %s, want 1088a980-eae5-11d0-8d9b-00a02453c337", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the (empty) transport endpoint: qm2qm is ncacn_ip_tcp only.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}

// TestOpnums verifies the opnums and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumRemoteQMStartReceive != 0 || OpnumRemoteQMGetQMQMServerPort != 7 || OpnumRemoteQMStartReceiveByLookupId != 10 {
		t.Fatalf("opnums = %d/%d/%d, want 0/7/10", OpnumRemoteQMStartReceive, OpnumRemoteQMGetQMQMServerPort, OpnumRemoteQMStartReceiveByLookupId)
	}
	if OpnumToName[0] != "RemoteQMStartReceive" || NameToOpnum["RemoteQmGetVersion"] != 8 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 11 {
		t.Fatalf("on-the-wire method count = %d, want 11", len(OpnumToName))
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
	if got := StatusString(STATUS_INVALID_PARAMETER); got != "STATUS_INVALID_PARAMETER" {
		t.Fatalf("StatusString(STATUS_INVALID_PARAMETER) = %s, want STATUS_INVALID_PARAMETER", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
