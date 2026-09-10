package rpcinterface_e33c0cc40482101abc0c02608c6ba218_1_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// TestSyntaxID verifies the abstract syntax identity (UUID + version) of the LocToLoc interface.
func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if got := s.UUID.ToFormatD(); got != "e33c0cc4-0482-101a-bc0c-02608c6ba218" {
		t.Fatalf("UUID = %s, want e33c0cc4-0482-101a-bc0c-02608c6ba218", got)
	}
	if s.MajorVersion != 1 || s.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", s.MajorVersion, s.MinorVersion)
	}
}

// TestPipeName verifies the transport endpoint ([MS-RPCL] section 2.1: \pipe\Locator).
func TestPipeName(t *testing.T) {
	if PipeName != `\Locator` {
		t.Fatalf("PipeName = %q, want %q", PipeName, `\Locator`)
	}
}

// TestOpnums verifies the seven on-the-wire opnums (RPC opnum order, [MS-RPCL] 3.1.4) and
// the name mapping.
func TestOpnums(t *testing.T) {
	want := map[uint16]string{
		0: "I_nsi_lookup_begin",
		1: "I_nsi_lookup_done",
		2: "I_nsi_lookup_next",
		3: "I_nsi_entry_object_inq_next",
		4: "I_nsi_ping_locator",
		5: "I_nsi_entry_object_inq_done",
		6: "I_nsi_entry_object_inq_begin",
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

// TestStatusString verifies mnemonic rendering and the hex fallback.
func TestStatusString(t *testing.T) {
	if got := StatusString(NSI_S_OK); got != "NSI_S_OK" {
		t.Fatalf("StatusString(NSI_S_OK) = %s, want NSI_S_OK", got)
	}
	if got := StatusString(NSI_S_NO_MORE_BINDINGS); got != "NSI_S_NO_MORE_BINDINGS" {
		t.Fatalf("StatusString(NSI_S_NO_MORE_BINDINGS) = %s, want NSI_S_NO_MORE_BINDINGS", got)
	}
	if got := StatusString(0xDEADBEEF); got != "0xdeadbeef" {
		t.Fatalf("StatusString(unknown) = %s, want 0xdeadbeef", got)
	}
}

// TestStatusCodesAreNSINotWin32 pins this interface's status codes to the Name Service
// Interface space they come from, so that a later pass does not route them through
// windows/errors/win32. [MS-RPCL] 3.1.4.3 documents NSI_S_OK as 0x00000000 and
// NSI_S_NO_MORE_BINDINGS as 0x00000001; [MS-ERREF] 2.2 gives that second value an
// unrelated name, and names the condition NSI_S_NO_MORE_BINDINGS reports under a
// different value altogether.
func TestStatusCodesAreNSINotWin32(t *testing.T) {
	if NSI_S_OK != 0x00000000 {
		t.Errorf("NSI_S_OK = 0x%08x, want 0x00000000 ([MS-RPCL] 3.1.4.3)", NSI_S_OK)
	}
	if NSI_S_NO_MORE_BINDINGS != 0x00000001 {
		t.Errorf("NSI_S_NO_MORE_BINDINGS = 0x%08x, want 0x00000001 ([MS-RPCL] 3.1.4.3)", NSI_S_NO_MORE_BINDINGS)
	}
	if StatusSuccess != NSI_S_OK {
		t.Errorf("StatusSuccess = 0x%08x, want NSI_S_OK 0x%08x", StatusSuccess, NSI_S_OK)
	}

	// The [MS-ERREF] 2.2 table names neither NSI code, so it cannot render either of
	// them; a migration to win32 would have nothing to migrate them to.
	for _, name := range []string{"NSI_S_OK", "NSI_S_NO_MORE_BINDINGS"} {
		if code, defined := win32.FromName(name); defined {
			t.Errorf("win32.FromName(%q) resolved to 0x%08x; %s is an NSI code and the Win32 table must not claim it", name, uint32(code), name)
		}
	}

	// What the Win32 table does say about 0x00000001, and why reporting the locator's
	// status through it would be wrong.
	if entry, defined := win32.Lookup(win32.WIN32_ERROR(NSI_S_NO_MORE_BINDINGS)); !defined {
		t.Errorf("win32 defines no name for 0x%08x; [MS-ERREF] 2.2 gives it ERROR_INVALID_FUNCTION", NSI_S_NO_MORE_BINDINGS)
	} else if entry.Name != "ERROR_INVALID_FUNCTION" {
		t.Errorf("win32 names 0x%08x %q, want ERROR_INVALID_FUNCTION: the collision this interface must keep out of its status reporting", NSI_S_NO_MORE_BINDINGS, entry.Name)
	}

	// The Win32 code for "there are no more bindings" is a different value entirely, so
	// the two spaces do not even agree where they name the same condition.
	if code, defined := win32.FromName("RPC_S_NO_MORE_BINDINGS"); !defined {
		t.Error(`win32.FromName("RPC_S_NO_MORE_BINDINGS") is undefined, want 0x0000070e ([MS-ERREF] 2.2)`)
	} else if uint32(code) != 0x0000070E || uint32(code) == NSI_S_NO_MORE_BINDINGS {
		t.Errorf("RPC_S_NO_MORE_BINDINGS = 0x%08x, want 0x0000070e and not NSI_S_NO_MORE_BINDINGS 0x%08x", uint32(code), NSI_S_NO_MORE_BINDINGS)
	}

	// StatusString renders only the NSI names and never falls through to the Win32
	// table: 0x00000002 is ERROR_FILE_NOT_FOUND there and has no NSI meaning here.
	if got := StatusString(0x00000002); got != "0x00000002" {
		t.Errorf("StatusString(0x00000002) = %s, want 0x00000002: the NSI status field must not be named out of the Win32 table", got)
	}
}
