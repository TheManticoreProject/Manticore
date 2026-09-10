package rpcinterface_41208ee0e97011d19b9e00e02c064c39_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the qmmgmt interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "41208ee0-e970-11d1-9b9e-00e02c064c39" {
		t.Fatalf("UUID = %s, want 41208ee0-e970-11d1-9b9e-00e02c064c39", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the (empty) transport endpoint: qmmgmt is ncacn_ip_tcp only.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}

// TestOpnums verifies the two on-the-wire opnums and the name mapping round-trip.
func TestOpnums(t *testing.T) {
	if OpnumR_QMMgmtGetInfo != 0 || OpnumR_QMMgmtAction != 1 {
		t.Fatalf("opnums = %d/%d, want 0/1", OpnumR_QMMgmtGetInfo, OpnumR_QMMgmtAction)
	}
	if OpnumToName[0] != "R_QMMgmtGetInfo" || NameToOpnum["R_QMMgmtAction"] != 1 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
	if len(OpnumToName) != 2 {
		t.Fatalf("on-the-wire method count = %d, want 2", len(OpnumToName))
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
		MQ_ERROR:                   "MQ_ERROR",
		MQ_ERROR_INVALID_PARAMETER: "MQ_ERROR_INVALID_PARAMETER",
		MQ_ERROR_ILLEGAL_PROPID:    "MQ_ERROR_ILLEGAL_PROPID",
	}
	if len(retained) != 3 {
		t.Fatalf("retained code count = %d, want 3", len(retained))
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
