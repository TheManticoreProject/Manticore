package ebcdic_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/encoding/ebcdic"
	"golang.org/x/text/encoding/charmap"
)

// allBytes returns a fresh slice holding 0x00..0xFF in order.
func allBytes() []uint8 {
	b := make([]uint8, 256)
	for i := range b {
		b[i] = uint8(i)
	}
	return b
}

// TestEbcdicToAsciiCp037MatchesXText decodes every EBCDIC byte with
// EbcdicToAscii("cp037") and checks it against the independent
// golang.org/x/text CodePage037 decoder, so a table bound in the wrong
// direction cannot pass.
func TestEbcdicToAsciiCp037MatchesXText(t *testing.T) {
	got, err := ebcdic.EbcdicToAscii("cp037", allBytes())
	if err != nil {
		t.Fatalf("EbcdicToAscii: %v", err)
	}
	for b := 0; b < 256; b++ {
		want := charmap.CodePage037.DecodeByte(byte(b))
		if rune(got[b]) != want {
			t.Errorf("cp037 EBCDIC 0x%02X -> 0x%02X, want %U", b, got[b], want)
		}
	}
}

// TestAsciiToEbcdicCp037InvertsXText checks that AsciiToEbcdic("cp037")
// maps every Latin-1 byte to the EBCDIC byte x/text decodes back to it.
func TestAsciiToEbcdicCp037InvertsXText(t *testing.T) {
	got, err := ebcdic.AsciiToEbcdic("cp037", allBytes())
	if err != nil {
		t.Fatalf("AsciiToEbcdic: %v", err)
	}
	for c := 0; c < 256; c++ {
		if r := charmap.CodePage037.DecodeByte(got[c]); r != rune(c) {
			t.Errorf("cp037 Latin-1 0x%02X -> EBCDIC 0x%02X which decodes to %U", c, got[c], r)
		}
	}
	enc, _ := ebcdic.AsciiToEbcdic("cp037", []uint8("A 0"))
	if string(enc) != "\xC1\x40\xF0" {
		t.Errorf(`AsciiToEbcdic("cp037", "A 0") = % X, want C1 40 F0`, enc)
	}
}

// cp500Anchors are code points from the IBM code page 500 chart. 0x4A, 0x4F,
// 0x5A, 0xBA and 0xBB are where CP500 differs from CP037.
var cp500Anchors = map[uint8]uint8{
	0x40: ' ',
	0x4B: '.',
	0x81: 'a',
	0xC1: 'A',
	0xF0: '0',
	0xF9: '9',
	0x4A: '[',
	0x4F: '!',
	0x5A: ']',
	0xBA: 0xAC, // NOT SIGN
	0xBB: '|',
}

// TestEbcdicToAsciiCp500Anchors decodes known CP500 bytes and checks the
// encoder inverts them.
func TestEbcdicToAsciiCp500Anchors(t *testing.T) {
	for e, want := range cp500Anchors {
		got, err := ebcdic.EbcdicToAscii("cp500", []uint8{e})
		if err != nil {
			t.Fatalf("EbcdicToAscii: %v", err)
		}
		if got[0] != want {
			t.Errorf("cp500 EBCDIC 0x%02X -> 0x%02X, want 0x%02X", e, got[0], want)
		}
		back, err := ebcdic.AsciiToEbcdic("cp500", []uint8{want})
		if err != nil {
			t.Fatalf("AsciiToEbcdic: %v", err)
		}
		if back[0] != e {
			t.Errorf("cp500 Latin-1 0x%02X -> 0x%02X, want 0x%02X", want, back[0], e)
		}
	}
}

// TestCodepageTablesAreMutualInverses checks that, for every code page, the
// two direction tables are bijective inverses of each other.
func TestCodepageTablesAreMutualInverses(t *testing.T) {
	for _, cp := range ebcdic.AvailableCodepages {
		e2a := ebcdic.CodepageToEbcdicToAsciiMap[cp]
		a2e := ebcdic.CodepageToAsciiToEbcdicMap[cp]
		if len(e2a) != 256 || len(a2e) != 256 {
			t.Fatalf("%s: table lengths %d/%d, want 256", cp, len(e2a), len(a2e))
		}
		for i := 0; i < 256; i++ {
			if e2a[a2e[i]] != uint8(i) {
				t.Errorf("%s: e2a[a2e[0x%02X]] = 0x%02X", cp, i, e2a[a2e[i]])
			}
			if a2e[e2a[i]] != uint8(i) {
				t.Errorf("%s: a2e[e2a[0x%02X]] = 0x%02X", cp, i, a2e[e2a[i]])
			}
		}
	}
}
