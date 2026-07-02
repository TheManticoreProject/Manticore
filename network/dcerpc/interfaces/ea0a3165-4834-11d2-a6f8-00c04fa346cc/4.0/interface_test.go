package rpcinterface_ea0a3165483411d2a6f800c04fa346cc_4_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// TestSyntaxID verifies the abstract syntax identifier matches the IDL: UUID
// ea0a3165-4834-11d2-a6f8-00c04fa346cc, version 4.0 ([MS-FAX]).
func TestSyntaxID(t *testing.T) {
	sid := SyntaxID()
	want := guid.GUID{A: 0xea0a3165, B: 0x4834, C: 0x11d2, D: 0xa6f8, E: 0x00c04fa346cc}
	if sid.UUID != want {
		t.Errorf("UUID = %+v, want %+v", sid.UUID, want)
	}
	if sid.MajorVersion != 4 || sid.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 4.0", sid.MajorVersion, sid.MinorVersion)
	}
}

// TestOpnumNameRoundTrip confirms OpnumToName and its derived reverse map agree, and that
// the wire opnums are the dense range 0..103 the fax interface defines.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 104 {
		t.Fatalf("OpnumToName has %d entries, want 104", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, want %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d (ok=%v), want %d", name, got, ok, op)
		}
	}
	for op := uint16(0); op < 104; op++ {
		if _, ok := OpnumToName[op]; !ok {
			t.Errorf("opnum %d missing from OpnumToName", op)
		}
	}
}

// TestStatusString spot-checks the mnemonic table and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:         "ERROR_SUCCESS",
		ErrorAccessDenied:     "ERROR_ACCESS_DENIED",
		ErrorInvalidParameter: "ERROR_INVALID_PARAMETER",
		FaxErrGroupNotFound:   "FAX_ERR_GROUP_NOT_FOUND",
		FaxErrRecipientsLimit: "FAX_ERR_RECIPIENTS_LIMIT",
		0xDEADBEEF:            "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
}
