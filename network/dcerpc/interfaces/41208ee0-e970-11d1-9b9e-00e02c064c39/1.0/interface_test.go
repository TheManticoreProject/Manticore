package rpcinterface_41208ee0e97011d19b9e00e02c064c39_1_0

import "testing"

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

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "MQ_OK" {
		t.Fatalf("StatusString(StatusSuccess) = %s, want MQ_OK", got)
	}
	if got := StatusString(MQ_ERROR_INVALID_PARAMETER); got != "MQ_ERROR_INVALID_PARAMETER" {
		t.Fatalf("StatusString(MQ_ERROR_INVALID_PARAMETER) = %s, want MQ_ERROR_INVALID_PARAMETER", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
