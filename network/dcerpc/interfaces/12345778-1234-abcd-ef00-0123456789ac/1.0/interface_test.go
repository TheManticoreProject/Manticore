package rpcinterface_123457781234abcdef000123456789ac_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// TestReturnCodesResolveThroughNTStatus pins that the status codes [MS-SAMR]
// documents for these methods resolve through the shared [MS-ERREF] table, which
// is what the method stubs in functions render with now that this package carries
// no subset of its own.
func TestReturnCodesResolveThroughNTStatus(t *testing.T) {
	cases := map[nt_status.NT_STATUS]string{
		nt_status.NT_STATUS_SUCCESS:               "NT_STATUS_SUCCESS",
		nt_status.NT_STATUS_MORE_ENTRIES:          "NT_STATUS_MORE_ENTRIES",
		nt_status.NT_STATUS_SOME_NOT_MAPPED:       "NT_STATUS_SOME_NOT_MAPPED",
		nt_status.NT_STATUS_NO_MORE_ENTRIES:       "NT_STATUS_NO_MORE_ENTRIES",
		nt_status.NT_STATUS_INVALID_HANDLE:        "NT_STATUS_INVALID_HANDLE",
		nt_status.NT_STATUS_INVALID_PARAMETER:     "NT_STATUS_INVALID_PARAMETER",
		nt_status.NT_STATUS_ACCESS_DENIED:         "NT_STATUS_ACCESS_DENIED",
		nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND: "NT_STATUS_OBJECT_NAME_NOT_FOUND",
		nt_status.NT_STATUS_NO_SUCH_USER:          "NT_STATUS_NO_SUCH_USER",
		nt_status.NT_STATUS_NONE_MAPPED:           "NT_STATUS_NONE_MAPPED",
		nt_status.NT_STATUS_NO_SUCH_DOMAIN:        "NT_STATUS_NO_SUCH_DOMAIN",
		nt_status.NT_STATUS_NO_SUCH_ALIAS:         "NT_STATUS_NO_SUCH_ALIAS",
		nt_status.NT_STATUS_NO_SUCH_GROUP:         "NT_STATUS_NO_SUCH_GROUP",
		nt_status.NT_STATUS_USER_EXISTS:           "NT_STATUS_USER_EXISTS",
		nt_status.NT_STATUS_GROUP_EXISTS:          "NT_STATUS_GROUP_EXISTS",
		nt_status.NT_STATUS_ALIAS_EXISTS:          "NT_STATUS_ALIAS_EXISTS",
		nt_status.NT_STATUS_WRONG_PASSWORD:        "NT_STATUS_WRONG_PASSWORD",
		nt_status.NT_STATUS_NOT_SUPPORTED:         "NT_STATUS_NOT_SUPPORTED",
	}

	for status, want := range cases {
		if got := status.String(); got != want {
			t.Errorf("0x%08X renders as %q, want %q", uint32(status), got, want)
		}
	}

	// A status this interface has never returned still renders, where the
	// package's own subset reported an undecoded hex value for anything outside
	// the 18 codes above.
	if got := nt_status.NT_STATUS_LOGON_FAILURE.String(); got != "NT_STATUS_LOGON_FAILURE" {
		t.Errorf("a status outside the old subset renders as %q", got)
	}
	if got := nt_status.NT_STATUS(0x12345678).String(); got != "0x12345678" {
		t.Errorf("an undefined status renders as %q, want the hex form", got)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if id.UUID.A != 0x12345778 || id.UUID.B != 0x1234 || id.UUID.C != 0xabcd ||
		id.UUID.D != 0xef00 || id.UUID.E != 0x0123456789ac {
		t.Errorf("SyntaxID UUID = %+v, want 12345778-1234-abcd-ef00-0123456789ac", id.UUID)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 64 {
		t.Errorf("OpnumToName has %d entries, want 64 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	if OpnumToName[OpnumSamrConnect5] != "SamrConnect5" {
		t.Errorf("OpnumToName[64] = %q", OpnumToName[OpnumSamrConnect5])
	}
}
