package rpcinterface_a8e0653c27444389a61d7373df8b2292_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// TestSyntaxID confirms the abstract syntax is the FileServerVssAgent UUID/version
// (a8e0653c-2744-4389-a61d-7373df8b2292 v1.0, [MS-FSRVP] 2.1).
func TestSyntaxID(t *testing.T) {
	want, err := guid.FromString("a8e0653c-2744-4389-a61d-7373df8b2292")
	if err != nil {
		t.Fatalf("FromString: %v", err)
	}
	id := SyntaxID()
	if id.UUID != *want {
		t.Errorf("UUID = %s, want %s", id.UUID.ToFormatD(), want.ToFormatD())
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

// TestOpnumNameRoundTrip verifies OpnumToName covers all 13 on-the-wire methods and that
// NameToOpnum is a faithful inverse.
func TestOpnumNameRoundTrip(t *testing.T) {
	if len(OpnumToName) != 13 {
		t.Fatalf("OpnumToName has %d entries, want 13", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, want %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d (ok=%v), want %d", name, got, ok, op)
		}
	}
}

// TestStatusString spot-checks the known-code mnemonics and the hex fallback.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:                 "ZERO",
		FsrvpEBadState:                "FSRVP_E_BAD_STATE",
		FsrvpEShadowcopysetIdMismatch: "FSRVP_E_SHADOWCOPYSET_ID_MISMATCH",
		FsrvpEWaitFailed:              "FSRVP_E_WAIT_FAILED",
		0xdeadbeef:                    "0xdeadbeef",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", code, got, want)
		}
	}
}
