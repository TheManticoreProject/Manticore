package msbpau

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestKEY_LENGTH_RoundTrip exercises the KEY_LENGTH scalar typedef ([MS-BPAU] 2.2.1):
// a 4-byte DWORD. It is wrapped in a struct because the NDR codec marshals struct
// values, and confirms the named type marshals with the same width as a bare DWORD.
func TestKEY_LENGTH_RoundTrip(t *testing.T) {
	type holder struct{ Len KEY_LENGTH }

	for _, v := range []KEY_LENGTH{0, 1, 4096, KeyLengthMax} {
		raw, err := ndr.Marshal(&holder{Len: v})
		if err != nil {
			t.Fatalf("Marshal(%d): %v", v, err)
		}
		if len(raw) != 4 {
			t.Fatalf("KEY_LENGTH marshalled to %d bytes, want 4 (DWORD width)", len(raw))
		}
		var out holder
		if err := ndr.Unmarshal(raw, &out); err != nil {
			t.Fatalf("Unmarshal(%d): %v", v, err)
		}
		if out.Len != v {
			t.Fatalf("KEY_LENGTH round-trip: got %d want %d", out.Len, v)
		}
	}
}
