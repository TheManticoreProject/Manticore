package rpcinterface_44e265dd7daf42cd85603cdb6e7a2729_1_3

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the TsProxyRpcInterface
// (44e265dd-7daf-42cd-8560-3cdb6e7a2729 v1.3, [MS-TSGU]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "44e265dd-7daf-42cd-8560-3cdb6e7a2729" {
		t.Errorf("UUID = %s, want 44e265dd-7daf-42cd-8560-3cdb6e7a2729", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 3 {
		t.Errorf("version = %d.%d, want 1.3", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses so the
// two maps never drift, and that only the 8 on-the-wire opnums are present (0 and 5 are
// "not used on the wire").
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 8 {
		t.Fatalf("OpnumToName has %d entries, want 8", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for _, op := range []uint16{1, 2, 3, 4, 6, 7, 8, 9} {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
	// Opnums 0 and 5 are not used on the wire and must be absent.
	for _, op := range []uint16{0, 5} {
		if _, ok := OpnumToName[op]; ok {
			t.Errorf("opnum %d should be absent (NotUsedOnWire)", op)
		}
	}
}

// TestStatusString checks a couple of mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(E_PROXY_INTERNALERROR); got != "E_PROXY_INTERNALERROR" {
		t.Errorf("StatusString(0x800759D8) = %q, want E_PROXY_INTERNALERROR", got)
	}
	if got := StatusString(E_PROXY_TS_CONNECTFAILED_CODE); got != "E_PROXY_TS_CONNECTFAILED" {
		t.Errorf("StatusString(0x59DD) = %q, want E_PROXY_TS_CONNECTFAILED", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}
