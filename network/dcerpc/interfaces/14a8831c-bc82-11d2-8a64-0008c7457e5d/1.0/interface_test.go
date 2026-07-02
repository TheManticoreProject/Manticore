package rpcinterface_14a8831cbc8211d28a640008c7457e5d_1_0

import "testing"

// TestSyntaxID checks the abstract syntax identifier for the ExtendedError interface
// ([MS-EERR]): UUID 14a8831c-bc82-11d2-8a64-0008c7457e5d, version 1.0.
func TestSyntaxID(t *testing.T) {
	id := SyntaxID()
	if got := id.UUID.ToFormatD(); got != "14a8831c-bc82-11d2-8a64-0008c7457e5d" {
		t.Fatalf("UUID = %s, want 14a8831c-bc82-11d2-8a64-0008c7457e5d", got)
	}
	if id.MajorVersion != 1 || id.MinorVersion != 0 {
		t.Fatalf("version = %d.%d, want 1.0", id.MajorVersion, id.MinorVersion)
	}
}

// TestNoOpnums confirms [MS-EERR] defines no on-the-wire methods: both opnum maps are
// empty and remain mutual inverses.
func TestNoOpnums(t *testing.T) {
	if len(OpnumToName) != 0 {
		t.Fatalf("OpnumToName has %d entries, want 0", len(OpnumToName))
	}
	if len(NameToOpnum) != len(OpnumToName) {
		t.Fatalf("NameToOpnum has %d entries, want %d", len(NameToOpnum), len(OpnumToName))
	}
}

// TestPipeNameEmpty documents that the ExtendedError interface has no named-pipe
// endpoint: its structures are pickled and embedded in other protocols' responses.
func TestPipeNameEmpty(t *testing.T) {
	if PipeName != "" {
		t.Fatalf("PipeName = %q, want empty (interface is not bound over a pipe)", PipeName)
	}
}
