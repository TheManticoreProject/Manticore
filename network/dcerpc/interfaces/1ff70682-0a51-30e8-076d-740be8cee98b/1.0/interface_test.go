package rpcinterface_1ff706820a5130e8076d740be8cee98b_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of atsvc.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "1ff70682-0a51-30e8-076d-740be8cee98b" {
		t.Fatalf("UUID = %s, want 1ff70682-0a51-30e8-076d-740be8cee98b", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName verifies the well-known ncacn_np endpoint shared with SASec.
func TestPipeName(t *testing.T) {
	if PipeName != `\atsvc` {
		t.Fatalf("PipeName = %q, want \\atsvc", PipeName)
	}
}

// TestOpnums verifies the implemented opnums and the name mapping round trip.
func TestOpnums(t *testing.T) {
	if OpnumNetrJobAdd != 0 || OpnumNetrJobDel != 1 || OpnumNetrJobEnum != 2 || OpnumNetrJobGetInfo != 3 {
		t.Fatalf("opnums = %d/%d/%d/%d, want 0/1/2/3", OpnumNetrJobAdd, OpnumNetrJobDel, OpnumNetrJobEnum, OpnumNetrJobGetInfo)
	}
	if OpnumToName[0] != "NetrJobAdd" || NameToOpnum["NetrJobGetInfo"] != 3 {
		t.Fatal("opnum name mapping is inconsistent")
	}
	if len(OpnumToName) != len(NameToOpnum) {
		t.Fatalf("OpnumToName (%d) and NameToOpnum (%d) disagree on size", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughWin32 pins that the seven NET_API_STATUS codes
// [MS-TSCH] 3.2.5.2 documents for ATSvc resolve through the shared [MS-ERREF] 2.2 table
// under their specification names, that codes the interface never enumerated now render
// by name rather than as undecoded hex, and that a value the specification does not
// define still renders as hex.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000002: "ERROR_FILE_NOT_FOUND",
		0x00000005: "ERROR_ACCESS_DENIED",
		0x00000032: "ERROR_NOT_SUPPORTED",
		0x00000057: "ERROR_INVALID_PARAMETER",
		0x0000007C: "ERROR_INVALID_LEVEL",
		0x000000EA: "ERROR_MORE_DATA",
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

	// These sit outside the subset the descriptor used to declare, so a method returning
	// one of them used to render as undecoded hex. They resolve by name now.
	outsideOldSubset := map[uint32]string{
		0x00000006: "ERROR_INVALID_HANDLE",
		0x00000008: "ERROR_NOT_ENOUGH_MEMORY",
		0x0000007B: "ERROR_INVALID_NAME",
		0x00000103: "ERROR_NO_MORE_ITEMS",
		0x00000424: "ERROR_SERVICE_DOES_NOT_EXIST",
	}
	for code, name := range outsideOldSubset {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}
