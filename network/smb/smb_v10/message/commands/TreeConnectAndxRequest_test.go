package commands_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// TestTreeConnectAndxRequestUnmarshalFreshNoPanic verifies that calling
// Unmarshal on a request constructed with NewTreeConnectAndxRequest does not
// panic with a nil-pointer dereference. The constructor does not initialize the
// Parameters/Data structures, so Unmarshal must create them when nil (mirroring
// Marshal) before using them.
func TestTreeConnectAndxRequestUnmarshalFreshNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unmarshal panicked on a freshly constructed request: %v", r)
		}
	}()

	c := commands.NewTreeConnectAndxRequest()

	// A minimal SMB_Parameters/SMB_Data byte stream: WordCount=0 then ByteCount=0.
	// The exact decoding is irrelevant here; the point is that Unmarshal must not
	// panic on the nil Parameters/Data structures of a fresh request.
	input := []byte{0x00, 0x00, 0x00}

	if _, err := c.Unmarshal(input); err != nil {
		// An error is acceptable for this minimal input; a panic is not.
		t.Logf("Unmarshal returned an error (acceptable, no panic): %v", err)
	}
}
