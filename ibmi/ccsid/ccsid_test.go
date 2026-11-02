package ccsid_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/encoding/ebcdic"
	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
)

func mustLookup(t *testing.T, id ccsid.CCSID) ccsid.Codec {
	t.Helper()
	c, err := ccsid.Lookup(id)
	if err != nil {
		t.Fatalf("Lookup(%d): %v", id, err)
	}
	if c.CCSID() != id {
		t.Fatalf("Lookup(%d).CCSID() = %d", id, c.CCSID())
	}
	return c
}

func TestLookup(t *testing.T) {
	for _, id := range ccsid.Supported() {
		mustLookup(t, id)
	}
	for _, id := range []ccsid.CCSID{0, 930, 5035, 65535} {
		c, err := ccsid.Lookup(id)
		var ue *ccsid.UnsupportedError
		if c != nil || !errors.Is(err, ccsid.ErrUnsupported) || !errors.As(err, &ue) || ue.CCSID != id {
			t.Errorf("Lookup(%d) = %v, %v; want UnsupportedError", id, c, err)
		}
	}
}

func TestCurrencyDiffers(t *testing.T) {
	c37, c1140 := mustLookup(t, ccsid.CCSID37), mustLookup(t, ccsid.CCSID1140)
	s37, _ := c37.Decode([]byte{0x9F})
	s1140, _ := c1140.Decode([]byte{0x9F})
	if s37 != "¤" || s1140 != "€" {
		t.Errorf("0x9F decodes to %q (37) and %q (1140)", s37, s1140)
	}
	if _, err := c37.Encode("€"); !errors.Is(err, ebcdic.ErrUnmappable) {
		t.Errorf("CCSID 37 Encode(€) error = %v", err)
	}
}

func TestEncodeFixed(t *testing.T) {
	c := mustLookup(t, ccsid.CCSID37)
	got, err := ccsid.EncodeFixed(c, "QGY", 10)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0xD8, 0xC7, 0xE8, 0x40, 0x40, 0x40, 0x40, 0x40, 0x40, 0x40}
	if !bytes.Equal(got, want) {
		t.Errorf("EncodeFixed = % X, want % X", got, want)
	}
	if got, err := ccsid.EncodeFixed(c, "ABCDEFGHIJ", 10); err != nil || len(got) != 10 {
		t.Errorf("exact-width EncodeFixed = % X, %v", got, err)
	}
	if got, err := ccsid.EncodeFixed(c, "", 0); err != nil || len(got) != 0 {
		t.Errorf("zero-width EncodeFixed = % X, %v", got, err)
	}
	if _, err := ccsid.EncodeFixed(c, "ABCDEFGHIJK", 10); !errors.Is(err, ccsid.ErrFieldTooLong) {
		t.Errorf("overlong EncodeFixed error = %v", err)
	}
	if _, err := ccsid.EncodeFixed(c, "€", 10); !errors.Is(err, ebcdic.ErrUnmappable) {
		t.Errorf("unmappable EncodeFixed error = %v", err)
	}
	if _, err := ccsid.EncodeFixed(c, "A", -1); err == nil {
		t.Error("negative width accepted")
	}
}

func TestDecodeFixed(t *testing.T) {
	c := mustLookup(t, ccsid.CCSID37)
	tests := []struct {
		in   []byte
		want string
	}{
		{[]byte{0xD8, 0xC7, 0xE8, 0x40, 0x40}, "QGY"},
		{[]byte{0x40, 0xC1, 0x40}, " A"},        // leading blank is data
		{[]byte{0x40, 0x40}, ""},                // all blanks
		{[]byte{}, ""},                          // empty field
		{[]byte{0xC1, 0x00}, "A\x00"},           // NUL is data, not padding
		{[]byte{0xC1, 0x40, 0xC2, 0x40}, "A B"}, // embedded blank kept
	}
	for _, tt := range tests {
		orig := append([]byte(nil), tt.in...)
		got, err := ccsid.DecodeFixed(c, tt.in)
		if err != nil || got != tt.want {
			t.Errorf("DecodeFixed(% X) = %q, %v; want %q", tt.in, got, err, tt.want)
		}
		if !bytes.Equal(tt.in, orig) {
			t.Errorf("DecodeFixed modified its input")
		}
	}
}

// TestNationalCodePages decodes characters whose position differs between the
// national code pages, as IBM's mappings place them, and checks every
// supported codec maps all 256 bytes to distinct runes and back.
func TestNationalCodePages(t *testing.T) {
	anchors := []struct {
		id   ccsid.CCSID
		b    byte
		want string
	}{
		{ccsid.CCSID273, 0x4A, "Ä"}, {ccsid.CCSID273, 0x7C, "§"},
		{ccsid.CCSID277, 0x7B, "Æ"}, {ccsid.CCSID278, 0x7B, "Ä"},
		{ccsid.CCSID278, 0x71, "\\"}, {ccsid.CCSID278, 0xE0, "É"},
		{ccsid.CCSID280, 0x7B, "£"}, {ccsid.CCSID284, 0x7B, "Ñ"},
		{ccsid.CCSID285, 0x5B, "£"}, {ccsid.CCSID285, 0xA1, "¯"},
		{ccsid.CCSID297, 0x7C, "à"}, {ccsid.CCSID297, 0x44, "@"},
		{ccsid.CCSID500, 0x4A, "["}, {ccsid.CCSID871, 0x4A, "Þ"},
		{ccsid.CCSID871, 0xC0, "þ"}, {ccsid.CCSID1047, 0xAD, "["},
		{ccsid.CCSID1141, 0x9F, "€"}, {ccsid.CCSID1142, 0x5A, "€"},
		{ccsid.CCSID1143, 0x5A, "€"}, {ccsid.CCSID1144, 0x9F, "€"},
		{ccsid.CCSID1145, 0x9F, "€"}, {ccsid.CCSID1146, 0x9F, "€"},
		{ccsid.CCSID1147, 0x9F, "€"}, {ccsid.CCSID1148, 0x9F, "€"},
		{ccsid.CCSID1149, 0x9F, "€"},
	}
	for _, a := range anchors {
		got, err := mustLookup(t, a.id).Decode([]byte{a.b})
		if err != nil || got != a.want {
			t.Errorf("CCSID %d byte 0x%02X = %q, %v; want %q", a.id, a.b, got, err, a.want)
		}
	}

	for _, id := range ccsid.Supported() {
		c := mustLookup(t, id)
		all := make([]byte, 256)
		for i := range all {
			all[i] = byte(i)
		}
		text, err := c.Decode(all)
		if err != nil {
			t.Fatalf("CCSID %d: %v", id, err)
		}
		if n := len([]rune(text)); n != 256 {
			t.Errorf("CCSID %d decodes 256 bytes to %d runes", id, n)
		}
		back, err := c.Encode(text)
		if err != nil || string(back) != string(all) {
			t.Errorf("CCSID %d does not round-trip all 256 bytes: %v", id, err)
		}
		if blank, _ := c.Decode([]byte{ccsid.Blank}); blank != " " {
			t.Errorf("CCSID %d: 0x40 decodes to %q, not a space", id, blank)
		}
	}

	// A rune outside the code page is refused, not substituted.
	if _, err := mustLookup(t, ccsid.CCSID297).Encode("€"); err == nil {
		t.Error("CCSID 297 encoded the euro sign, which only 1147 carries")
	}
}
