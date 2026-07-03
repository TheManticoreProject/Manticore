package rpcinterface_76d12b80346711d391ff0090272f9ea3_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the qmcomm2 interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "76d12b80-3467-11d3-91ff-0090272f9ea3" {
		t.Fatalf("UUID = %s, want 76d12b80-3467-11d3-91ff-0090272f9ea3", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the opnums and the name mapping.
func TestOpnums(t *testing.T) {
	if OpnumQMSendMessageInternalEx != 0 || Opnumrpc_ACReceiveMessageEx != 2 || Opnumrpc_ACCreateCursorEx != 3 {
		t.Fatalf("opnums = %d/%d/%d, want 0/2/3", OpnumQMSendMessageInternalEx, Opnumrpc_ACReceiveMessageEx, Opnumrpc_ACCreateCursorEx)
	}
	if OpnumToName[0] != "QMSendMessageInternalEx" || NameToOpnum["rpc_ACCreateCursorEx"] != 3 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 4 {
		t.Fatalf("on-the-wire method count = %d, want 4", len(OpnumToName))
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
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
