package ndr

import (
	"bytes"
	"testing"
)

func TestEncoder_Alignment(t *testing.T) {
	e := NewEncoder()
	e.WriteUint8(0x01)
	e.WriteUint32(0x02) // must 4-align: 3 pad bytes after the uint8
	want := []byte{0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}
	if !bytes.Equal(e.Bytes(), want) {
		t.Errorf("aligned encode:\n got %x\nwant %x", e.Bytes(), want)
	}
}

func TestEncoder_Uint64Alignment(t *testing.T) {
	e := NewEncoder()
	e.WriteUint16(0xaabb)
	e.WriteUint64(0x1122334455667788) // 8-align: 6 pad bytes after the uint16
	want := []byte{
		0xbb, 0xaa, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11,
	}
	if !bytes.Equal(e.Bytes(), want) {
		t.Errorf("got %x want %x", e.Bytes(), want)
	}
}

func TestWString_GoldenBytes(t *testing.T) {
	e := NewEncoder()
	e.writeWString("AB")
	// count = 3 (incl NUL); offset 0; actual 3; "A\0" "B\0" little-endian; NUL terminator.
	want := []byte{
		0x03, 0x00, 0x00, 0x00, // maximum_count
		0x00, 0x00, 0x00, 0x00, // offset
		0x03, 0x00, 0x00, 0x00, // actual_count
		0x41, 0x00, 0x42, 0x00, // "AB"
		0x00, 0x00, // terminator
	}
	if !bytes.Equal(e.Bytes(), want) {
		t.Errorf("wstring:\n got %x\nwant %x", e.Bytes(), want)
	}
}

func TestWString_RoundTrip(t *testing.T) {
	for _, s := range []string{"", "A", `\\host\share\file.txt`, "Ünîçødé"} {
		e := NewEncoder()
		e.writeWString(s)
		got, err := NewDecoder(e.Bytes()).readWString()
		if err != nil {
			t.Fatalf("readWString(%q): %v", s, err)
		}
		if got != s {
			t.Errorf("round trip: got %q want %q", got, s)
		}
	}
}

func TestConformantBytes_RoundTrip(t *testing.T) {
	in := []byte{0xaa, 0xbb, 0xcc}
	e := NewEncoder()
	e.writeConformantBytes(in)
	want := []byte{0x03, 0x00, 0x00, 0x00, 0xaa, 0xbb, 0xcc}
	if !bytes.Equal(e.Bytes(), want) {
		t.Fatalf("conformant bytes:\n got %x\nwant %x", e.Bytes(), want)
	}
	got, err := NewDecoder(e.Bytes()).readConformantBytes()
	if err != nil {
		t.Fatalf("readConformantBytes: %v", err)
	}
	if !bytes.Equal(got, in) {
		t.Errorf("round trip: got %x want %x", got, in)
	}
}

func TestDecoder_BoundsChecked(t *testing.T) {
	d := NewDecoder([]byte{0x01, 0x02})
	if _, err := d.ReadUint32(); err == nil {
		t.Fatal("ReadUint32 on a 2-byte stream: error = nil, want non-nil")
	}
}
