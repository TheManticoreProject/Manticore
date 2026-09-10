package rpcinterface_da5a86c512c24943ab307f74a813d853_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

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

// TestStatusCodesResolveThroughWin32 pins that the seven Win32 codes [MS-PCQ] sections
// 3.1.4.1-3.1.4.8 document for PerflibV2 resolve through the shared [MS-ERREF] 2.2 table
// under their specification names, that codes the interface never enumerated now render
// by name rather than as undecoded hex, and that a value the specification does not
// define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	// [MS-ERREF] 2.2 carries two names on the 0x00000000 row, ERROR_SUCCESS and the
	// NERR_Success alias from the network-management range; String renders the canonical
	// ERROR_SUCCESS, which is the name the descriptor's StatusString returned.
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x00001068: "ERROR_WMI_GUID_NOT_FOUND",
		0x0000106A: "ERROR_WMI_ITEMID_NOT_FOUND",
	}
	if len(documented) != 7 {
		t.Fatalf("documented table has %d entries, want the 7 codes the descriptor declared", len(documented))
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// These sit outside the subset the descriptor used to declare. ERROR_WMI_INSTANCE_NOT_FOUND
	// falls between the two WMI codes it did declare and ERROR_WMI_TRY_AGAIN just past them, so
	// a counter query that named an unknown instance or raced the provider used to render as
	// undecoded hex. They resolve by name now.
	outsideOldSubset := map[uint32]string{
		0x0000000D: "ERROR_INVALID_DATA",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x0000007A: "ERROR_INSUFFICIENT_BUFFER",
		0x000000EA: "ERROR_MORE_DATA",
		0x00001069: "ERROR_WMI_INSTANCE_NOT_FOUND",
		0x0000106B: "ERROR_WMI_TRY_AGAIN",
	}
	for code, name := range outsideOldSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}
