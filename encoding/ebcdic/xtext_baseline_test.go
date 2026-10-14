package ebcdic_test

import (
	"testing"

	"golang.org/x/text/encoding/charmap"
)

// TestXTextCP037CP1140Baseline pins the reviewed golang.org/x/text tables that
// the IBM i CCSID 37/1140 codec wraps: both are 256-byte bijections, they
// differ only at 0x9F (currency sign in CP037, euro sign in CP1140), and the
// EBCDIC blank used for fixed-width IBM i fields is 0x40.
func TestXTextCP037CP1140Baseline(t *testing.T) {
	for _, tc := range []struct {
		name string
		cm   *charmap.Charmap
	}{
		{"CP037", charmap.CodePage037},
		{"CP1140", charmap.CodePage1140},
	} {
		seen := map[rune]byte{}
		for i := 0; i < 256; i++ {
			b := byte(i)
			r := tc.cm.DecodeByte(b)
			if prev, dup := seen[r]; dup {
				t.Fatalf("%s: bytes 0x%02X and 0x%02X both decode to %U", tc.name, prev, b, r)
			}
			seen[r] = b
			back, ok := tc.cm.EncodeRune(r)
			if !ok || back != b {
				t.Fatalf("%s: 0x%02X -> %U -> 0x%02X (ok=%v)", tc.name, b, r, back, ok)
			}
		}
		if r := tc.cm.DecodeByte(0x40); r != ' ' {
			t.Errorf("%s: 0x40 decodes to %U, want space", tc.name, r)
		}
		if r := tc.cm.DecodeByte(0xC1); r != 'A' {
			t.Errorf("%s: 0xC1 decodes to %U, want 'A'", tc.name, r)
		}
	}

	for i := 0; i < 256; i++ {
		b := byte(i)
		r037, r1140 := charmap.CodePage037.DecodeByte(b), charmap.CodePage1140.DecodeByte(b)
		switch {
		case b == 0x9F:
			if r037 != '¤' || r1140 != '€' {
				t.Errorf("0x9F: CP037=%U CP1140=%U, want U+00A4 and U+20AC", r037, r1140)
			}
		case r037 != r1140:
			t.Errorf("0x%02X: CP037=%U CP1140=%U, tables must differ only at 0x9F", b, r037, r1140)
		}
	}
}
