package rpcinterface_e33c0cc40482101abc0c02608c6ba218_1_0

import "testing"

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
