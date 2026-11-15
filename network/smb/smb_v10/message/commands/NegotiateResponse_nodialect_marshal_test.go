package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// Test_NegotiateResponse_Marshal_NoDialectSupported checks that a no-common-dialect
// refusal marshals to the minimal [MS-CIFS] 2.2.4.52.2 form: WordCount 1 carrying
// only DialectIndex 0xFFFF, with ByteCount 0. Emitting the full 17-word structure
// with zeroed fields is a server fingerprint, since Windows and Samba send only
// the one word.
func Test_NegotiateResponse_Marshal_NoDialectSupported(t *testing.T) {
	resp := commands.NewNegotiateResponse()
	resp.DialectIndex = commands.NoDialectSupported

	got, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// WordCount 0x01, DialectIndex 0xFFFF (LE), ByteCount 0x0000.
	want := []byte{0x01, 0xFF, 0xFF, 0x00, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("no-dialect marshal = % x, want % x", got, want)
	}
}

// Test_NegotiateResponse_Marshal_SelectedDialectUnaffected guards the normal path:
// a response that selected a dialect still marshals the full NT LM 0.12 body
// (WordCount 17), not the minimal refusal.
func Test_NegotiateResponse_Marshal_SelectedDialectUnaffected(t *testing.T) {
	resp := commands.NewNegotiateResponse()
	resp.DialectIndex = 0

	got, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("empty marshal")
	}
	if got[0] != 17 {
		t.Fatalf("selected-dialect WordCount = %d, want 17", got[0])
	}
}
