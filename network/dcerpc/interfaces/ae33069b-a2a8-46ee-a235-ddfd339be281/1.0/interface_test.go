package rpcinterface_ae33069ba2a846eea235ddfd339be281_1_0

import "testing"

// TestSyntaxID checks the abstract syntax UUID and version match [MS-PAN] Appendix A.2.
func TestSyntaxID(t *testing.T) {
	sid := SyntaxID()
	if got := sid.UUID.ToFormatD(); got != "ae33069b-a2a8-46ee-a235-ddfd339be281" {
		t.Errorf("UUID = %s, want ae33069b-a2a8-46ee-a235-ddfd339be281", got)
	}
	if sid.MajorVersion != 1 || sid.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", sid.MajorVersion, sid.MinorVersion)
	}
}

// TestPipeName pins the (empty) transport endpoint: MS-PAN is ncacn_ip_tcp only.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (ncacn_ip_tcp dynamic endpoint)", PipeName)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent inverses.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName has %d entries, NameToOpnum has %d", len(OpnumToName), len(NameToOpnum))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d (ok=%v), want %d", name, got, ok, op)
		}
	}
}

// TestStatusString spot-checks the HRESULT mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:     "S_OK",
		ErrorAccessDenied: "E_ACCESSDENIED",
		0xdeadbeef:        "0xdeadbeef",
	}
	for status, want := range cases {
		if got := StatusString(status); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", status, got, want)
		}
	}
}
