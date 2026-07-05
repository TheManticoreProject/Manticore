package rpcinterface_d95afe70a6d54259822e2c84da1ddb0d_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the WindowsShutdown interface
// (d95afe70-a6d5-4259-822e-2c84da1ddb0d v1.0, [MS-RSP]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "d95afe70-a6d5-4259-822e-2c84da1ddb0d" {
		t.Errorf("UUID = %s, want d95afe70-a6d5-4259-822e-2c84da1ddb0d", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover exactly the 2 on-the-wire opnums (0, 1).
func TestOpnumNameRoundTrip(t *testing.T) {
	wire := []uint16{0, 1}
	if len(OpnumToName) != len(wire) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(wire))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("round trip failed: opnum %d -> %q -> %d", op, name, NameToOpnum[name])
		}
	}
	for _, op := range wire {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusString checks known mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:           "ERROR_SUCCESS",
		ErrorAccessDenied:       "ERROR_ACCESS_DENIED",
		ErrorShutdownInProgress: "ERROR_SHUTDOWN_IN_PROGRESS",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the empty endpoint: WindowsShutdown is ncacn_ip_tcp / dynamic
// endpoint per [MS-RSP] 2.1, so it has no well-known named pipe.
func TestPipeName(t *testing.T) {
	if PipeName != `` {
		t.Errorf("PipeName = %q, want empty (TCP dynamic endpoint)", PipeName)
	}
}
