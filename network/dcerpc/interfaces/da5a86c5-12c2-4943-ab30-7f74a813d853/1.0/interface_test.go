package rpcinterface_da5a86c512c24943ab307f74a813d853_1_0

import "testing"

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the PerflibV2
// interface ([MS-PCQ] Appendix A).
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "da5a86c5-12c2-4943-ab30-7f74a813d853" {
		t.Fatalf("UUID = %s, want da5a86c5-12c2-4943-ab30-7f74a813d853", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName verifies the ncacn_np well-known endpoint ([MS-PCQ] 2.1).
func TestPipeName(t *testing.T) {
	if PipeName != `\winreg` {
		t.Fatalf("PipeName = %q, want %q", PipeName, `\winreg`)
	}
}

// TestOpnums verifies the eight on-the-wire opnums and the name mapping round-trip.
func TestOpnums(t *testing.T) {
	want := map[uint16]string{
		0: "PerflibV2EnumerateCounterSet",
		1: "PerflibV2QueryCounterSetRegistrationInfo",
		2: "PerflibV2EnumerateCounterSetInstances",
		3: "PerflibV2OpenQueryHandle",
		4: "PerflibV2CloseQueryHandle",
		5: "PerflibV2QueryCounterInfo",
		6: "PerflibV2QueryCounterData",
		7: "PerflibV2ValidateCounters",
	}
	if len(OpnumToName) != len(want) {
		t.Fatalf("OpnumToName has %d entries, want %d", len(OpnumToName), len(want))
	}
	for op, name := range want {
		if OpnumToName[op] != name {
			t.Fatalf("OpnumToName[%d] = %q, want %q", op, OpnumToName[op], name)
		}
		if NameToOpnum[name] != op {
			t.Fatalf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusString verifies mnemonic rendering of the documented Win32 error codes and the
// hex fallback for unknown codes.
func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		StatusSuccess:          "ERROR_SUCCESS",
		ErrorAccessDenied:      "ERROR_ACCESS_DENIED",
		ErrorInvalidHandle:     "ERROR_INVALID_HANDLE",
		ErrorNotEnoughMemory:   "ERROR_NOT_ENOUGH_MEMORY",
		ErrorInvalidParameter:  "ERROR_INVALID_PARAMETER",
		ErrorWmiGuidNotFound:   "ERROR_WMI_GUID_NOT_FOUND",
		ErrorWmiItemIdNotFound: "ERROR_WMI_ITEMID_NOT_FOUND",
	}
	for code, want := range cases {
		if got := StatusString(code); got != want {
			t.Fatalf("StatusString(0x%08x) = %s, want %s", code, got, want)
		}
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}
