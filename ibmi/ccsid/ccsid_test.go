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
	for _, id := range []ccsid.CCSID{0, 273, 500, 1047, 65535} {
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
