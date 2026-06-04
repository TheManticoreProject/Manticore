package ndr

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

func TestWriterReaderRoundTrip(t *testing.T) {
	g := guid.GUID{A: 0x11223344, B: 0x5566, C: 0x7788, D: 0x99aa, E: 0xbbccddeeff00}

	var w Writer
	w.U32(0xdeadbeef)
	w.Put([]byte{0xaa})
	w.Align(4) // pad from length 5 to 8
	w.Put(g.ToBytes())
	w.U32(0x01020304)

	if got := len(w.Bytes()); got != 4+1+3+16+4 {
		t.Fatalf("stream length = %d, want %d", got, 4+1+3+16+4)
	}

	r := NewReader(w.Bytes())
	if v := r.U32(); v != 0xdeadbeef {
		t.Errorf("U32 = 0x%08x, want 0xdeadbeef", v)
	}
	if b := r.Take(1); len(b) != 1 || b[0] != 0xaa {
		t.Errorf("Take(1) = % x, want aa", b)
	}
	r.Align(4)
	if got := r.UUID(); !got.Equal(&g) {
		t.Errorf("UUID = %s, want %s", got.ToFormatD(), g.ToFormatD())
	}
	if v := r.U32(); v != 0x01020304 {
		t.Errorf("trailing U32 = 0x%08x, want 0x01020304", v)
	}
	if err := r.Err(); err != nil {
		t.Fatalf("Err after full read: %v", err)
	}
}

func TestReaderUnderrunIsSticky(t *testing.T) {
	r := NewReader([]byte{0x01, 0x02})
	if v := r.U32(); v != 0 { // only 2 bytes available
		t.Errorf("U32 on short buffer = %d, want 0", v)
	}
	if r.Err() == nil {
		t.Fatal("expected a sticky underrun error")
	}
	// Once failed, further reads stay zero and the error persists.
	if v := r.U16(); v != 0 {
		t.Errorf("U16 after error = %d, want 0", v)
	}
	if r.Err() == nil {
		t.Fatal("error should remain set")
	}
}

func TestReaderAlignUnderrun(t *testing.T) {
	r := NewReader([]byte{0x01}) // 1 byte
	_ = r.Take(1)                // off = 1
	r.Align(4)                   // cannot advance to 4: underrun
	if r.Err() == nil {
		t.Fatal("expected underrun error aligning past the end")
	}
}
