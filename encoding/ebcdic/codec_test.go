package ebcdic_test

import (
	"bytes"
	"errors"
	"testing"
	"unicode/utf8"

	"github.com/TheManticoreProject/Manticore/encoding/ebcdic"
)

// cp037Anchors are code points of IBM code page 037 taken from the IBM
// CCSID 37 code chart, independent of the table under test.
var cp037Anchors = map[byte]rune{
	0x00: 0x0000, 0x40: ' ', 0x4B: '.', 0x4C: '<', 0x4D: '(', 0x4E: '+', 0x4F: '|',
	0x50: '&', 0x5A: '!', 0x5B: '$', 0x5C: '*', 0x5D: ')', 0x5E: ';', 0x5F: '¬',
	0x60: '-', 0x61: '/', 0x6B: ',', 0x6C: '%', 0x6D: '_', 0x6E: '>', 0x6F: '?',
	0x79: '`', 0x7A: ':', 0x7B: '#', 0x7C: '@', 0x7D: '\'', 0x7E: '=', 0x7F: '"',
	0xB0: '^', 0xBA: '[', 0xBB: ']', 0xC0: '{', 0xD0: '}', 0xE0: '\\', 0xA1: '~',
	0x9F: '¤',
}

func init() {
	for i := 0; i < 9; i++ {
		cp037Anchors[byte(0xC1+i)] = rune('A' + i)
		cp037Anchors[byte(0xD1+i)] = rune('J' + i)
		cp037Anchors[byte(0x81+i)] = rune('a' + i)
		cp037Anchors[byte(0x91+i)] = rune('j' + i)
	}
	for i := 0; i < 8; i++ {
		cp037Anchors[byte(0xE2+i)] = rune('S' + i)
		cp037Anchors[byte(0xA2+i)] = rune('s' + i)
	}
	for i := 0; i < 10; i++ {
		cp037Anchors[byte(0xF0+i)] = rune('0' + i)
	}
}

func decodeAll(t *testing.T, c *ebcdic.Codec) []rune {
	t.Helper()
	all := make([]byte, 256)
	for i := range all {
		all[i] = byte(i)
	}
	s, err := c.Decode(all)
	if err != nil {
		t.Fatalf("%s: Decode: %v", c.Name(), err)
	}
	runes := []rune(s)
	if len(runes) != 256 {
		t.Fatalf("%s: decoded %d runes, want 256", c.Name(), len(runes))
	}
	return runes
}

// TestCodecDecodesEveryByte checks all 256 bytes of both code pages: CP037
// is a permutation of ISO 8859-1 (U+0000..U+00FF), and CP1140 differs from
// it only at 0x9F, where the currency sign becomes the euro sign.
func TestCodecDecodesEveryByte(t *testing.T) {
	r037 := decodeAll(t, ebcdic.CP037())
	r1140 := decodeAll(t, ebcdic.CP1140())

	seen := map[rune]bool{}
	for b, r := range r037 {
		if r > 0xFF {
			t.Errorf("cp037 0x%02X -> %U, outside ISO 8859-1", b, r)
		}
		if seen[r] {
			t.Errorf("cp037 %U decoded twice", r)
		}
		seen[r] = true
		if want, ok := cp037Anchors[byte(b)]; ok && r != want {
			t.Errorf("cp037 0x%02X -> %U, want %U", b, r, want)
		}
	}
	for b := range r1140 {
		switch {
		case b == 0x9F:
			if r1140[b] != '€' {
				t.Errorf("cp1140 0x9F -> %U, want U+20AC", r1140[b])
			}
		case r1140[b] != r037[b]:
			t.Errorf("0x%02X: cp1140 %U differs from cp037 %U", b, r1140[b], r037[b])
		}
	}
}

func TestCodecRoundTrip(t *testing.T) {
	for _, c := range []*ebcdic.Codec{ebcdic.CP037(), ebcdic.CP1140()} {
		for _, s := range []string{"", "QSYS", "Profile description: café, naïve, Ærø!", "a{b}[c]\\d|e~f"} {
			enc, err := c.Encode(s)
			if err != nil {
				t.Fatalf("%s Encode(%q): %v", c.Name(), s, err)
			}
			if len(enc) != utf8.RuneCountInString(s) {
				t.Errorf("%s Encode(%q) produced %d bytes, want one per rune", c.Name(), s, len(enc))
			}
			dec, _ := c.Decode(enc)
			if dec != s {
				t.Errorf("%s round trip %q -> %q", c.Name(), s, dec)
			}
		}
	}
	got, _ := ebcdic.CP037().Encode("A 0")
	if !bytes.Equal(got, []byte{0xC1, 0x40, 0xF0}) {
		t.Errorf("cp037 Encode(\"A 0\") = % X", got)
	}
}

func TestCodecRejectsUnmappable(t *testing.T) {
	tests := []struct {
		codec  *ebcdic.Codec
		input  string
		rune   rune
		offset int
	}{
		{ebcdic.CP037(), "price 5€", '€', 7},
		{ebcdic.CP1140(), "x¤", '¤', 1},
		{ebcdic.CP037(), "名", '名', 0},
		{ebcdic.CP037(), "ok\xffbad", utf8.RuneError, 2},
	}
	for _, tt := range tests {
		out, err := tt.codec.Encode(tt.input)
		if out != nil || !errors.Is(err, ebcdic.ErrUnmappable) {
			t.Fatalf("%s Encode(%q) = %v, %v; want ErrUnmappable", tt.codec.Name(), tt.input, out, err)
		}
		var ue *ebcdic.UnmappableError
		if !errors.As(err, &ue) || ue.Rune != tt.rune || ue.Offset != tt.offset || ue.Codepage != tt.codec.Name() {
			t.Errorf("%s Encode(%q) error = %#v", tt.codec.Name(), tt.input, ue)
		}
	}
	// A literal U+FFFD is valid UTF-8 but not in either repertoire.
	if _, err := ebcdic.CP037().Encode("�"); !errors.Is(err, ebcdic.ErrUnmappable) {
		t.Errorf("Encode(U+FFFD) error = %v", err)
	}
}

func TestCodecDoesNotModifyInput(t *testing.T) {
	in := []byte{0xC1, 0xC2, 0x40, 0x9F}
	orig := append([]byte(nil), in...)
	if _, err := ebcdic.CP1140().Decode(in); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(in, orig) {
		t.Fatalf("Decode modified its input: % X", in)
	}
	a, _ := ebcdic.CP037().Encode("AB")
	b, _ := ebcdic.CP037().Encode("AB")
	a[0] = 0
	if b[0] != 0xC1 {
		t.Fatal("Encode results share storage")
	}
}
