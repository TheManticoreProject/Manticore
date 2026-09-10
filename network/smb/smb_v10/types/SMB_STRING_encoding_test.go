package types

import (
	"testing"
)

// TestSetStringWithEncodingRoundTrip checks that a string stored in either
// encoding reads back unchanged. SetString stores the UTF-8 bytes of a Go string,
// which is right for OEM and wrong for Unicode: the wire wants UTF-16LE there, and
// a caller that stores UTF-8 sends a different name than it asked for.
func TestSetStringWithEncodingRoundTrip(t *testing.T) {
	for _, name := range []string{"plain.txt", "unicode_éàü", `\dir\nested_ß.txt`, ""} {
		for _, unicode := range []bool{false, true} {
			s := &SMB_STRING{}
			if err := s.SetStringWithEncoding(name, unicode); err != nil {
				t.Fatalf("SetStringWithEncoding(%q, %v): %v", name, unicode, err)
			}
			if got := s.StringWithEncoding(unicode); got != name {
				t.Errorf("unicode=%v: round-tripped %q as %q", unicode, name, got)
			}
			if int(s.Length) != len(s.Buffer) {
				t.Errorf("unicode=%v: Length %d does not match the %d buffer bytes", unicode, s.Length, len(s.Buffer))
			}
		}
	}
}

// TestSetStringWithEncodingWidth checks the Unicode form really is two bytes per
// ASCII character — a length check is what distinguishes UTF-16LE from the UTF-8
// bytes SetString would have stored.
func TestSetStringWithEncodingWidth(t *testing.T) {
	const ascii = "abc"

	oem := &SMB_STRING{}
	if err := oem.SetStringWithEncoding(ascii, false); err != nil {
		t.Fatalf("SetStringWithEncoding: %v", err)
	}
	if len(oem.Buffer) != 3 {
		t.Errorf("OEM buffer is %d bytes, want 3", len(oem.Buffer))
	}

	uni := &SMB_STRING{}
	if err := uni.SetStringWithEncoding(ascii, true); err != nil {
		t.Fatalf("SetStringWithEncoding: %v", err)
	}
	if len(uni.Buffer) != 6 {
		t.Errorf("Unicode buffer is %d bytes, want 6", len(uni.Buffer))
	}
	// Little-endian: each ASCII character is its byte followed by a zero.
	if uni.Buffer[0] != 'a' || uni.Buffer[1] != 0x00 {
		t.Errorf("Unicode buffer starts % x, want 61 00", uni.Buffer[:2])
	}
}
