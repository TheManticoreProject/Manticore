package rpcinterface_f6beaff71e194fbb9f8fb89e2018337c_1_0

import "testing"

// TestSyntaxID pins the abstract syntax: f6beaff7-1e19-4fbb-9f8f-b89e2018337c v1.0.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	u := s.UUID
	if u.A != 0xf6beaff7 || u.B != 0x1e19 || u.C != 0x4fbb || u.D != 0x9f8f || u.E != 0xb89e2018337c {
		t.Errorf("UUID = %s, want f6beaff7-1e19-4fbb-9f8f-b89e2018337c", u.ToFormatD())
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName pins the transport endpoint (RPC endpoint name "Eventlog", [MS-EVEN6]
// Standards Assignments).
func TestPipeName(t *testing.T) {
	if PipeName != `\eventlog` {
		t.Errorf("PipeName = %q, want %q", PipeName, `\eventlog`)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName and NameToOpnum are consistent and that all
// 29 on-the-wire opnums (0..28, contiguous) are present exactly once.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 29 {
		t.Fatalf("OpnumToName has %d entries, want 29", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, want %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	for op := uint16(0); op <= 28; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("on-the-wire opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusString spot-checks a few mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:              "ERROR_SUCCESS",
		ErrorAccessDenied:          "ERROR_ACCESS_DENIED",
		ErrorNoMoreItems:           "ERROR_NO_MORE_ITEMS",
		ErrorEvtInvalidChannelPath: "ERROR_EVT_INVALID_CHANNEL_PATH",
		ErrorEvtChannelNotFound:    "ERROR_EVT_CHANNEL_NOT_FOUND",
		ErrorEvtMessageIDNotFound:  "ERROR_EVT_MESSAGE_ID_NOT_FOUND",
		0xdeadbeef:                 "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
}
