package rpcinterface_6bffd098a11236109833012892020162_0_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "6bffd098-a112-3610-9833-012892020162" {
		t.Fatalf("UUID = %s, want 6bffd098-a112-3610-9833-012892020162", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the single on-the-wire opnum and the name mapping round-trip.
func TestOpnums(t *testing.T) {
	if OpnumI_BrowserrQueryOtherDomains != 2 {
		t.Fatalf("OpnumI_BrowserrQueryOtherDomains = %d, want 2", OpnumI_BrowserrQueryOtherDomains)
	}
	if OpnumToName[2] != "I_BrowserrQueryOtherDomains" || NameToOpnum["I_BrowserrQueryOtherDomains"] != 2 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName/NameToOpnum size mismatch: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(ERROR_INVALID_LEVEL); got != "ERROR_INVALID_LEVEL" {
		t.Fatalf("StatusString(ERROR_INVALID_LEVEL) = %s", got)
	}
	if got := StatusString(NERR_Success); got != "NERR_Success" {
		t.Fatalf("StatusString(NERR_Success) = %s", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the transport endpoint ([MS-BRWSA] 2.1).
func TestPipeName(t *testing.T) {
	if PipeName != `\browser` {
		t.Fatalf("PipeName = %q, want %q", PipeName, `\browser`)
	}
}
