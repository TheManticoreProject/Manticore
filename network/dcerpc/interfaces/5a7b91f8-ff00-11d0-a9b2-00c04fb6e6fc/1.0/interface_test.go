package rpcinterface_5a7b91f8ff0011d0a9b200c04fb6e6fc_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc" {
		t.Fatalf("UUID = %s, want 5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the single opnum and its name mapping round-trips.
func TestOpnums(t *testing.T) {
	if OpnumNetrSendMessage != 0 {
		t.Fatalf("OpnumNetrSendMessage = %d, want 0", OpnumNetrSendMessage)
	}
	if OpnumToName[0] != "NetrSendMessage" || NameToOpnum["NetrSendMessage"] != 0 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("map sizes differ: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering of documented codes and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:     "ERROR_SUCCESS",
		ErrorAccessDenied: "ERROR_ACCESS_DENIED",
		NerrNetworkError:  "NERR_NetworkError",
		NerrDuplicateName: "NERR_DuplicateName",
		0xdeadbeef:        "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
}
