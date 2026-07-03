package rpcinterface_0b6edbfa4a244fc68a23942b1eca65d1_1_0

import "testing"

// TestSyntaxID checks the abstract syntax UUID and version match [MS-PAN] Appendix A.1.
func TestSyntaxID(t *testing.T) {
	sid := SyntaxID()
	if got := sid.UUID.ToFormatD(); got != "0b6edbfa-4a24-4fc6-8a23-942b1eca65d1" {
		t.Errorf("UUID = %s, want 0b6edbfa-4a24-4fc6-8a23-942b1eca65d1", got)
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

// TestOpnums verifies the opnums, that Opnum2NotUsedOnWire is omitted, and the name map.
func TestOpnums(t *testing.T) {
	if OpnumIRPCAsyncNotify_RegisterClient != 0 || OpnumIRPCAsyncNotify_GetNewChannel != 3 || OpnumIRPCAsyncNotify_CloseChannel != 6 {
		t.Fatalf("opnums = %d/%d/%d, want 0/3/6", OpnumIRPCAsyncNotify_RegisterClient, OpnumIRPCAsyncNotify_GetNewChannel, OpnumIRPCAsyncNotify_CloseChannel)
	}
	// Opnum 2 (Opnum2NotUsedOnWire) MUST NOT appear in the map.
	if _, ok := OpnumToName[2]; ok {
		t.Error("opnum 2 (Opnum2NotUsedOnWire) must be omitted from OpnumToName")
	}
	if len(OpnumToName) != 6 {
		t.Fatalf("on-the-wire method count = %d, want 6", len(OpnumToName))
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

// TestStatusString spot-checks the protocol-specific HRESULT mnemonics and hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:     "S_OK",
		ErrorAccessDenied: "E_ACCESSDENIED",
		ErrorInvalidName:  "HRESULT_FROM_WIN32(ERROR_INVALID_NAME)",
		0xdeadbeef:        "0xdeadbeef",
	}
	for status, want := range cases {
		if got := StatusString(status); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", status, got, want)
		}
	}
}
