package rpcinterface_20610036fa2211cf982300a0c911e5df_1_0

import "testing"

// TestSyntaxID pins the abstract syntax identifier for the rasrpc interface
// (20610036-fa22-11cf-9823-00a0c911e5df v1.0, [MS-RRASM]).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "20610036-fa22-11cf-9823-00a0c911e5df" {
		t.Errorf("UUID = %s, want 20610036-fa22-11cf-9823-00a0c911e5df", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are exact inverses and
// cover exactly the 7 on-the-wire opnums (5, 9, 10, 11, 12, 14, 15). Opnums 0-4, 6-8,
// 13, 16 are "not used on the wire" and are intentionally absent.
func TestOpnumNameRoundTrip(t *testing.T) {
	wire := []uint16{5, 9, 10, 11, 12, 14, 15}
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
	for _, op := range []uint16{0, 1, 2, 3, 4, 6, 7, 8, 13, 16} {
		if _, ok := OpnumToName[op]; ok {
			t.Errorf("opnum %d is not used on the wire and must be absent", op)
		}
	}
}

// TestStatusString checks a known mnemonic and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(StatusSuccess); got != "ERROR_SUCCESS" {
		t.Errorf("StatusString(0) = %q, want ERROR_SUCCESS", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Errorf("StatusString(unknown) = %q, want 0xdeadbeef", got)
	}
}

// TestPipeName pins the [MS-RRASM] 2.1 well-known endpoint shared with dimsvc.
func TestPipeName(t *testing.T) {
	if PipeName != `\ROUTER` {
		t.Errorf("PipeName = %q, want \\ROUTER", PipeName)
	}
}
