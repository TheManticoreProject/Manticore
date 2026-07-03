package rpcinterface_fdb3a030065f11d1bb9b00a024ea5525_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the qmcomm interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "fdb3a030-065f-11d1-bb9b-00a024ea5525" {
		t.Fatalf("UUID = %s, want fdb3a030-065f-11d1-bb9b-00a024ea5525", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies representative opnums (including the gaps left by the "not used on
// the wire" methods) and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumR_QMGetRemoteQueueName != 1 || OpnumR_QMCreateRemoteCursor != 4 || OpnumR_QMCreateObjectInternal != 6 {
		t.Fatalf("opnums = %d/%d/%d, want 1/4/6", OpnumR_QMGetRemoteQueueName, OpnumR_QMCreateRemoteCursor, OpnumR_QMCreateObjectInternal)
	}
	if Opnumrpc_QMOpenQueueInternal != 19 || Opnumrpc_ACHandleToFormatName != 26 || OpnumR_QMGetRTQMServerPort != 31 {
		t.Fatalf("opnums = %d/%d/%d, want 19/26/31", Opnumrpc_QMOpenQueueInternal, Opnumrpc_ACHandleToFormatName, OpnumR_QMGetRTQMServerPort)
	}
	if OpnumToName[1] != "R_QMGetRemoteQueueName" || NameToOpnum["R_QMGetRTQMServerPort"] != 31 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 24 {
		t.Fatalf("on-the-wire method count = %d, want 24", len(OpnumToName))
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
