package guid

import "testing"

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
