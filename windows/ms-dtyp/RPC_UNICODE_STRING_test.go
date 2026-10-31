package msdtyp

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestNewUnicodeString_LongestEncodes checks that a string of exactly
// MaxUnicodeStringUnits code units gets the largest even byte counts and
// round-trips through NDR whole.
func TestNewUnicodeString_LongestEncodes(t *testing.T) {
	s := strings.Repeat("A", MaxUnicodeStringUnits)
	u := NewUnicodeString(s)
	if u.Length != 0xFFFE || u.MaximumLength != 0xFFFE || len(u.Buffer) != MaxUnicodeStringUnits {
		t.Fatalf("got Length=%d MaximumLength=%d len(Buffer)=%d, want 65534/65534/%d",
			u.Length, u.MaximumLength, len(u.Buffer), MaxUnicodeStringUnits)
	}
	raw, err := ndr.Marshal(&u)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out RPC_UNICODE_STRING
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.String() != s {
		t.Fatalf("round-trip length %d, want %d", len(out.String()), len(s))
	}
}

// TestNewUnicodeString_OverlongFailsToEncode checks that a string longer than the
// uint16 byte counts can describe is refused when encoded, instead of having its
// counts wrap or its data cut down to what the counts describe.
func TestNewUnicodeString_OverlongFailsToEncode(t *testing.T) {
	for _, n := range []int{MaxUnicodeStringUnits + 1, 40000, 70000} {
		u := NewUnicodeString(strings.Repeat("A", n))
		if len(u.Buffer) != n {
			t.Fatalf("n=%d: len(Buffer)=%d, want the whole string", n, len(u.Buffer))
		}
		if _, err := ndr.Marshal(&u); err == nil {
			t.Fatalf("n=%d: Marshal succeeded, want an error", n)
		}
	}
}
