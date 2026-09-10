package rpcinterface_76d12b80346711d391ff0090272f9ea3_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
)

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

// TestStatusStringSharedTable verifies that the statuses this interface no longer
// declares resolve through the shared [MS-ERREF] 2.1.1 table: MQ_OK is the HRESULT S_OK,
// and a generic HRESULT the interface never enumerated is named rather than rendered as
// hex.
func TestStatusStringSharedTable(t *testing.T) {
	if got := StatusString(uint32(hresult.S_OK)); got != "S_OK" {
		t.Errorf("StatusString(S_OK) = %s, want S_OK", got)
	}
	const eFail = 0x80004005
	if got := StatusString(eFail); got != "E_FAIL" {
		t.Errorf("StatusString(0x%08x) = %s, want E_FAIL", uint32(eFail), got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}

// TestStatusStringMSMQFacility verifies that every retained code still renders under its
// Message Queuing name, and that the shared table genuinely does not know it — [MS-ERREF]
// 2.1.1 carries no row in FACILITY_MSMQ, which is why these stay declared locally.
func TestStatusStringMSMQFacility(t *testing.T) {
	retained := map[uint32]string{
		MQ_ERROR:                          "MQ_ERROR",
		MQ_ERROR_QUEUE_NOT_FOUND:          "MQ_ERROR_QUEUE_NOT_FOUND",
		MQ_ERROR_QUEUE_NOT_ACTIVE:         "MQ_ERROR_QUEUE_NOT_ACTIVE",
		MQ_ERROR_INVALID_PARAMETER:        "MQ_ERROR_INVALID_PARAMETER",
		MQ_ERROR_INVALID_HANDLE:           "MQ_ERROR_INVALID_HANDLE",
		MQ_ERROR_OPERATION_CANCELLED:      "MQ_ERROR_OPERATION_CANCELLED",
		MQ_ERROR_SHARING_VIOLATION:        "MQ_ERROR_SHARING_VIOLATION",
		MQ_ERROR_SERVICE_NOT_AVAILABLE:    "MQ_ERROR_SERVICE_NOT_AVAILABLE",
		MQ_ERROR_MESSAGE_ALREADY_RECEIVED: "MQ_ERROR_MESSAGE_ALREADY_RECEIVED",
		MQ_ERROR_ACCESS_DENIED:            "MQ_ERROR_ACCESS_DENIED",
		MQ_ERROR_INSUFFICIENT_RESOURCES:   "MQ_ERROR_INSUFFICIENT_RESOURCES",
		MQ_ERROR_IO_TIMEOUT:               "MQ_ERROR_IO_TIMEOUT",
		MQ_ERROR_TRANSACTION_USAGE:        "MQ_ERROR_TRANSACTION_USAGE",
	}
	if len(retained) != 13 {
		t.Fatalf("retained code count = %d, want 13", len(retained))
	}
	for value, name := range retained {
		if got := StatusString(value); got != name {
			t.Errorf("StatusString(0x%08x) = %s, want %s", value, got, name)
		}
		if value>>16&0x07FF != 0x00E {
			t.Errorf("%s (0x%08x) is not in FACILITY_MSMQ", name, value)
		}
		if entry, defined := hresult.Lookup(hresult.HRESULT(value)); defined {
			t.Errorf("hresult.Lookup(0x%08x) = %s, want undefined", value, entry.Name)
		}
	}
}
