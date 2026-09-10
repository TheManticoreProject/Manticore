package rpcinterface_123457781234abcdef000123456789ab_0_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// TestDocumentedStatusesResolve pins that the NTSTATUS codes [MS-LSAD] and
// [MS-LSAT] list as return values of these methods resolve through the shared
// [MS-ERREF] 2.3.1 table under their NT_STATUS_* names, which is why this
// interface declares no subset of its own.
func TestDocumentedStatusesResolve(t *testing.T) {
	documented := map[nt_status.NT_STATUS]string{
		0x00000000: "NT_STATUS_SUCCESS",
		0x00000105: "NT_STATUS_MORE_ENTRIES",
		0x00000107: "NT_STATUS_SOME_NOT_MAPPED",
		0x8000001A: "NT_STATUS_NO_MORE_ENTRIES",
		0xC0000008: "NT_STATUS_INVALID_HANDLE",
		0xC000000D: "NT_STATUS_INVALID_PARAMETER",
		0xC0000022: "NT_STATUS_ACCESS_DENIED",
		0xC0000034: "NT_STATUS_OBJECT_NAME_NOT_FOUND",
		0xC0000060: "NT_STATUS_NO_SUCH_PRIVILEGE",
		0xC0000073: "NT_STATUS_NONE_MAPPED",
		0xC0000078: "NT_STATUS_INVALID_SID",
		0xC00000BB: "NT_STATUS_NOT_SUPPORTED",
		0xC00000DF: "NT_STATUS_NO_SUCH_DOMAIN",
	}
	for status, name := range documented {
		if got := status.String(); got != name {
			t.Errorf("NT_STATUS(0x%08x).String() = %q, want %q", uint32(status), got, name)
		}
	}
}

// TestStatusOutsideDocumentedSetResolves pins what the subset cost. A status
// these methods return but the subset never listed — LsarCreateAccount reports
// NT_STATUS_OBJECT_NAME_COLLISION when the account already exists — now renders
// by name where it used to render as undecoded hexadecimal. A value no
// specification defines still renders as hexadecimal, which is all that can be
// said about it.
func TestStatusOutsideDocumentedSetResolves(t *testing.T) {
	if got := nt_status.NT_STATUS(0xC0000035).String(); got != "NT_STATUS_OBJECT_NAME_COLLISION" {
		t.Errorf("NT_STATUS(0xc0000035).String() = %q, want %q", got, "NT_STATUS_OBJECT_NAME_COLLISION")
	}
	if got := nt_status.NT_STATUS(0x12345678).String(); got != "0x12345678" {
		t.Errorf("NT_STATUS(0x12345678).String() = %q, want the hexadecimal value", got)
	}
}

func TestOpnumNameMapsRoundTrip(t *testing.T) {
	if len(OpnumToName) != 69 {
		t.Errorf("OpnumToName has %d entries, want 69 on-the-wire methods", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d (a duplicate name collapsed an entry)",
			len(NameToOpnum), len(OpnumToName))
	}
	for op, name := range OpnumToName {
		if got, ok := NameToOpnum[name]; !ok || got != op {
			t.Errorf("NameToOpnum[%q] = %d, %v; want %d", name, got, ok, op)
		}
	}
	// Spot-check both directions.
	if OpnumToName[OpnumLsarOpenPolicy2] != "LsarOpenPolicy2" {
		t.Errorf("OpnumToName[44] = %q, want LsarOpenPolicy2", OpnumToName[OpnumLsarOpenPolicy2])
	}
	if NameToOpnum["LsarClose"] != OpnumLsarClose {
		t.Errorf("NameToOpnum[LsarClose] = %d, want %d", NameToOpnum["LsarClose"], OpnumLsarClose)
	}
}

func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	// 12345778-1234-abcd-ef00-0123456789ab, version 0.0.
	if id.UUID.A != 0x12345778 || id.UUID.B != 0x1234 || id.UUID.C != 0xabcd ||
		id.UUID.D != 0xef00 || id.UUID.E != 0x0123456789ab {
		t.Errorf("SyntaxID UUID = %+v, want 12345778-1234-abcd-ef00-0123456789ab", id.UUID)
	}
	if id.MajorVersion != 0 || id.MinorVersion != 0 {
		t.Errorf("SyntaxID version = %d.%d, want 0.0", id.MajorVersion, id.MinorVersion)
	}
}
