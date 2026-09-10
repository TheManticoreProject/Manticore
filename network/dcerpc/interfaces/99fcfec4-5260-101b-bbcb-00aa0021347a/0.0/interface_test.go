package rpcinterface_99fcfec45260101bbbcb00aa0021347a_0_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of IObjectExporter.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "99fcfec4-5260-101b-bbcb-00aa0021347a" {
		t.Fatalf("UUID = %s, want 99fcfec4-5260-101b-bbcb-00aa0021347a", got)
	}
	if s.MajorVersion != 0 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 0.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestOpnums verifies the six opnums and that OpnumToName/NameToOpnum stay in sync.
func TestOpnums(t *testing.T) {
	want := map[uint16]string{
		0: "ResolveOxid",
		1: "SimplePing",
		2: "ComplexPing",
		3: "ServerAlive",
		4: "ResolveOxid2",
		5: "ServerAlive2",
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
		t.Fatalf("map sizes differ: %d vs %d", len(OpnumToName), len(NameToOpnum))
	}
}

// TestStatusCodesResolveThroughWin32 pins that the four status codes this interface used
// to declare itself resolve through the shared [MS-ERREF] 2.2 table under their
// specification names, that a code the interface never enumerated now renders by name
// rather than as undecoded hex, and that a value the specification does not define still
// renders as hex. The object resolver failures [MS-DCOM] 3.1.2.5 documents are rows of
// that table, so the private subset was a copy of four of its 2703 codes.
func TestStatusCodesResolveThroughWin32(t *testing.T) {
	documented := map[uint32]string{
		0x00000000: "ERROR_SUCCESS",
		0x00000776: "OR_INVALID_OXID",
		0x00000777: "OR_INVALID_OID",
		0x00000778: "OR_INVALID_SET",
	}
	for code, name := range documented {
		if got := win32.WIN32_ERROR(code).String(); got != name {
			t.Errorf("win32.WIN32_ERROR(0x%08x).String() = %q, want %q", code, got, name)
		}
		if resolved, defined := win32.FromName(name); !defined || uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, %v; want 0x%08x, true", name, uint32(resolved), defined, code)
		}
	}

	// RPC_S_OK is the name [C706] gives the zero the table names ERROR_SUCCESS, and both
	// resolve to the same code, so reporting success through win32.ERROR_SUCCESS says what
	// the interface used to say with its own constant.
	if uint32(win32.ERROR_SUCCESS) != 0x00000000 {
		t.Errorf("win32.ERROR_SUCCESS = 0x%08x, want 0x00000000 (RPC_S_OK)", uint32(win32.ERROR_SUCCESS))
	}

	// RPC_S_SERVER_UNAVAILABLE, which a resolver that is not listening produces, was
	// outside the subset this interface used to declare, so it rendered as hex; it
	// resolves by name now.
	if got := win32.WIN32_ERROR(0x000006BA).String(); got != "RPC_S_SERVER_UNAVAILABLE" {
		t.Errorf("win32.WIN32_ERROR(0x000006ba).String() = %q, want RPC_S_SERVER_UNAVAILABLE", got)
	}

	// A value [MS-ERREF] 2.2 does not define still renders as hex.
	if got := win32.WIN32_ERROR(0xDEADBEEF).String(); got != "0xdeadbeef" {
		t.Errorf("win32.WIN32_ERROR(0xdeadbeef).String() = %q, want hex", got)
	}
}
