package rpcinterface_f5cc5a184264101a8c5908002b2f8426_56_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the nspi interface
// (f5cc5a18-4264-101a-8c59-08002b2f8426 v56.0, [MS-NSPI]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "f5cc5a18-4264-101a-8c59-08002b2f8426" {
		t.Errorf("UUID = %s, want f5cc5a18-4264-101a-8c59-08002b2f8426", got)
	}
	if s.MajorVersion != 56 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 56.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and that
// all 20 on-the-wire opnums are covered. Opnum 15 (Opnum15NotUsedOnWire) is intentionally
// absent, so the covered set is {0..14, 16..20}.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 20 {
		t.Fatalf("OpnumToName has %d entries, want 20", len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for op := uint16(0); op <= 20; op++ {
		_, ok := OpnumToName[op]
		if op == 15 {
			if ok {
				t.Errorf("opnum 15 is not used on the wire and must be absent")
			}
			continue
		}
		if !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusString checks a known success mnemonic, the ErrorsReturned warning, a failure
// code, and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:        "Success",
		StatusErrorsReturned: "ErrorsReturned",
		StatusLogonFailed:    "LogonFailed",
		0xDEADBEEF:           "0xdeadbeef",
	}
	for status, want := range cases {
		if got := StatusString(status); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", status, got, want)
		}
	}
}

// TestPipeName pins NSPI's empty pipe: it has no well-known named-pipe endpoint
// ([MS-NSPI] 2.1 uses dynamic endpoints across ncacn_np/ncacn_http/ncacn_ip_tcp).
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty", PipeName)
	}
}
