package rpcinterface_e1af83085d1f11c991a408002b14a0fa_3_0

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

func TestSyntaxID(t *testing.T) {
	s := SyntaxID()
	if s.MajorVersion != 3 || s.MinorVersion != 0 {
		t.Errorf("version = %d.%d, want 3.0", s.MajorVersion, s.MinorVersion)
	}
	want := "e1af8308-5d1f-11c9-91a4-08002b14a0fa"
	if got := s.UUID.ToFormatD(); got != want {
		t.Errorf("UUID = %s, want %s", got, want)
	}
}

func TestStatusString(t *testing.T) {
	cases := map[uint32]string{
		EptStatusSuccess:       "rpc_s_ok",
		EptStatusNotRegistered: "ept_s_not_registered",
		0xdeadbeef:             "0xdeadbeef",
	}
	for status, want := range cases {
		if got := StatusString(status); got != want {
			t.Errorf("StatusString(0x%08x) = %q, want %q", status, got, want)
		}
	}
}

func TestOpnumNameRoundTrip(t *testing.T) {
	for op, name := range OpnumToName {
		if NameToOpnum[name] != op {
			t.Errorf("NameToOpnum[%q] = %d, want %d", name, NameToOpnum[name], op)
		}
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Errorf("NameToOpnum has %d entries, OpnumToName has %d", len(NameToOpnum), len(OpnumToName))
	}
	if OpnumEptMap != 3 {
		t.Errorf("OpnumEptMap = %d, want 3", OpnumEptMap)
	}
}

// TestStatusCodesAreDCENotWin32 pins this interface's status codes to the DCE RPC space
// they come from, so a later pass does not route them through windows/errors/win32. The
// endpoint mapper's error_status_t carries values from the DCE status facility
// 0x16c9a0xx; [MS-ERREF] 2.2 has no row for any of them, and it names the same three
// conditions under Win32 values of its own.
func TestStatusCodesAreDCENotWin32(t *testing.T) {
	dce := map[string]uint32{
		"rpc_s_ok":              EptStatusSuccess,
		"ept_s_not_registered":  EptStatusNotRegistered,
		"ept_s_invalid_entry":   EptStatusInvalidEntry,
		"ept_s_cant_perform_op": EptStatusCantPerform,
	}
	want := map[string]uint32{
		"rpc_s_ok":              0x00000000,
		"ept_s_not_registered":  0x16c9a0d6,
		"ept_s_invalid_entry":   0x16c9a0d7,
		"ept_s_cant_perform_op": 0x16c9a0d8,
	}
	for name, code := range dce {
		if code != want[name] {
			t.Errorf("%s = 0x%08x, want 0x%08x", name, code, want[name])
		}
		// The shared table knows none of the DCE mnemonics, so a name-driven migration
		// would have nothing to resolve.
		if resolved, defined := win32.FromName(name); defined {
			t.Errorf("win32.FromName(%q) = 0x%08x, true; %s is a DCE code and the Win32 table must not claim it",
				name, uint32(resolved), name)
		}
	}

	// Nor does it have a row for any of the three ept_s_* values, so a value-driven
	// migration has nothing to resolve either.
	for _, code := range []uint32{EptStatusNotRegistered, EptStatusInvalidEntry, EptStatusCantPerform} {
		if entry, defined := win32.Lookup(win32.WIN32_ERROR(code)); defined {
			t.Errorf("win32 names 0x%08x %q; [MS-ERREF] 2.2 defines no code in the DCE 0x16c9a0xx facility",
				code, entry.Name)
		}
	}

	// The Win32 codes for the same three conditions, and how far they sit from the DCE
	// values this interface reads off the wire.
	parallel := map[string]uint32{
		"EPT_S_INVALID_ENTRY":   0x000006D7,
		"EPT_S_CANT_PERFORM_OP": 0x000006D8,
		"EPT_S_NOT_REGISTERED":  0x000006D9,
	}
	for name, code := range parallel {
		resolved, defined := win32.FromName(name)
		if !defined {
			t.Errorf("win32.FromName(%q) is undefined, want 0x%08x ([MS-ERREF] 2.2)", name, code)
			continue
		}
		if uint32(resolved) != code {
			t.Errorf("win32.FromName(%q) = 0x%08x, want 0x%08x", name, uint32(resolved), code)
		}
		for _, dceCode := range []uint32{EptStatusNotRegistered, EptStatusInvalidEntry, EptStatusCantPerform} {
			if uint32(resolved) == dceCode {
				t.Errorf("win32 %s = 0x%08x collides with a DCE value this interface declares", name, dceCode)
			}
		}
	}

	// The trap the parallel spaces set: two of the three DCE values share their low byte
	// with the Win32 code for the same condition, and ept_s_not_registered does not, so
	// transposing 0x16c9a0d6 into the Win32 block lands on an authorization error.
	if entry, defined := win32.Lookup(0x000006D6); !defined || entry.Name != "RPC_S_UNKNOWN_AUTHZ_SERVICE" {
		t.Errorf("win32.Lookup(0x000006d6) = %q, %v; want RPC_S_UNKNOWN_AUTHZ_SERVICE, the code a low-byte transposition of ept_s_not_registered would hit",
			entry.Name, defined)
	}

	// StatusString renders only the four DCE names and never falls through to the Win32
	// table, so neither a Win32 EPT_S_* value nor a small Win32 code is named out of it.
	for code, want := range map[uint32]string{0x000006D9: "0x000006d9", 0x00000002: "0x00000002"} {
		if got := StatusString(code); got != want {
			t.Errorf("StatusString(0x%08x) = %s, want %s: the DCE status field must not be named out of the Win32 table",
				code, got, want)
		}
	}
}
