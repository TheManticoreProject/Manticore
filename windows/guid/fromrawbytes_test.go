package guid

import (
	"encoding/binary"
	"testing"
)

// TestFromRawBytesOrder_BigEndian checks that the three integer UUID fields read in
// the given byte order while the clock_seq and node octets stay fixed. The same 16
// octets parse to byte-swapped A/B/C under big- vs little-endian, and D/E are equal.
func TestFromRawBytesOrder_BigEndian(t *testing.T) {
	raw := []byte{
		0x12, 0x34, 0x56, 0x78, // time_low
		0x9a, 0xbc, // time_mid
		0xde, 0xf0, // time_hi_and_version
		0x11, 0x22, // clock_seq (byte string)
		0x33, 0x44, 0x55, 0x66, 0x77, 0x88, // node (byte string)
	}

	var be GUID
	be.FromRawBytesOrder(raw, binary.BigEndian)
	if be.A != 0x12345678 || be.B != 0x9abc || be.C != 0xdef0 {
		t.Errorf("big-endian A/B/C = %#x/%#x/%#x, want 0x12345678/0x9abc/0xdef0", be.A, be.B, be.C)
	}

	var le GUID
	le.FromRawBytesOrder(raw, binary.LittleEndian)
	if le.A != 0x78563412 || le.B != 0xbc9a || le.C != 0xf0de {
		t.Errorf("little-endian A/B/C = %#x/%#x/%#x, want 0x78563412/0xbc9a/0xf0de", le.A, le.B, le.C)
	}

	// D (clock_seq) and E (node) are byte strings: identical under either order.
	if be.D != le.D || be.E != le.E {
		t.Errorf("byte-string fields differ by order: BE D/E = %#x/%#x, LE D/E = %#x/%#x", be.D, be.E, le.D, le.E)
	}
	if be.D != 0x1122 || be.E != 0x334455667788 {
		t.Errorf("D/E = %#x/%#x, want 0x1122/0x334455667788", be.D, be.E)
	}

	// FromRawBytes must stay the little-endian path.
	var def GUID
	def.FromRawBytes(raw)
	if !def.Equal(&le) {
		t.Errorf("FromRawBytes = %s, want the little-endian parse %s", def.ToFormatD(), le.ToFormatD())
	}
}

// TestFromRawBytes_ShortInput checks that FromRawBytes does not panic on fewer than
// 16 bytes and leaves the zero GUID, overwriting any previous value.
func TestFromRawBytes_ShortInput(t *testing.T) {
	for _, n := range []int{0, 1, 9, 15} {
		g := &GUID{A: 0x12345678, B: 0x1234, C: 0x5678, D: 0x9abc, E: 0xdef012345678}
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("FromRawBytes(%d bytes) panicked: %v", n, r)
				}
			}()
			g.FromRawBytes(make([]byte, n))
		}()
		if !g.Equal(&GUID{}) {
			t.Errorf("FromRawBytes(%d bytes): got %s, want the zero GUID", n, g.ToFormatD())
		}
	}
}

// TestFromRawBytes_LongInput checks that only the first 16 bytes are read.
func TestFromRawBytes_LongInput(t *testing.T) {
	want := &GUID{A: 0x12345678, B: 0x1234, C: 0x5678, D: 0x9abc, E: 0xdef012345678}
	g := &GUID{}
	g.FromRawBytes(append(want.ToBytes(), 0xff, 0xff))
	if !g.Equal(want) {
		t.Errorf("got %s, want %s", g.ToFormatD(), want.ToFormatD())
	}
}
