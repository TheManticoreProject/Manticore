package ndr

import (
	"bytes"
	"testing"
)

// TestSyntax_Default verifies NDR20 is the zero value, so a codec built without an
// explicit syntax behaves as the classic NDR 2.0 codec.
func TestSyntax_Default(t *testing.T) {
	if got := NewEncoder().Syntax(); got != NDR20 {
		t.Errorf("NewEncoder syntax = %v, want NDR20", got)
	}
	if got := NewDecoder(nil).Syntax(); got != NDR20 {
		t.Errorf("NewDecoder syntax = %v, want NDR20", got)
	}
	if got := NewEncoderForSyntax(NDR64).Syntax(); got != NDR64 {
		t.Errorf("NewEncoderForSyntax(NDR64) syntax = %v, want NDR64", got)
	}
	if NDR20.String() != "NDR20" || NDR64.String() != "NDR64" {
		t.Errorf("Syntax.String: got %q/%q, want NDR20/NDR64", NDR20.String(), NDR64.String())
	}
}

// TestWriteCount_NDR20Equivalence pins the count helper to the historical NDR20 wire
// form: a 4-octet little-endian value, 4-aligned, identical to a direct WriteUint32.
func TestWriteCount_NDR20Equivalence(t *testing.T) {
	via := NewEncoder()
	via.writeCount(0x11223344)

	direct := NewEncoder()
	direct.WriteUint32(0x11223344)

	if !bytes.Equal(via.Bytes(), direct.Bytes()) {
		t.Errorf("writeCount NDR20:\n got %x\nwant %x", via.Bytes(), direct.Bytes())
	}
	want := []byte{0x44, 0x33, 0x22, 0x11}
	if !bytes.Equal(via.Bytes(), want) {
		t.Errorf("writeCount NDR20 bytes:\n got %x\nwant %x", via.Bytes(), want)
	}
}

// TestWriteCount_NDR64 pins the count helper's NDR64 wire form: an 8-octet
// little-endian value, 8-aligned ([MS-RPCE] section 2.2.5).
func TestWriteCount_NDR64(t *testing.T) {
	e := NewEncoderForSyntax(NDR64)
	e.WriteUint8(0x01)       // force a non-8-aligned offset
	e.writeCount(0x11223344) // must 8-align: 7 pad bytes after the uint8
	want := []byte{
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // uint8 + 7 pad
		0x44, 0x33, 0x22, 0x11, 0x00, 0x00, 0x00, 0x00, // 8-octet count
	}
	if !bytes.Equal(e.Bytes(), want) {
		t.Errorf("writeCount NDR64:\n got %x\nwant %x", e.Bytes(), want)
	}
}

// TestCount_RoundTrip exercises the count helpers symmetrically in both syntaxes.
func TestCount_RoundTrip(t *testing.T) {
	for _, s := range []Syntax{NDR20, NDR64} {
		for _, v := range []uint64{0, 1, 0xffff, 0x7fffffff} {
			e := NewEncoderForSyntax(s)
			e.writeCount(v)
			got, err := NewDecoderForSyntax(e.Bytes(), s).readCount()
			if err != nil {
				t.Fatalf("%v readCount(%d): %v", s, v, err)
			}
			if got != v {
				t.Errorf("%v count round trip: got %d want %d", s, got, v)
			}
		}
	}
}

// TestReferent_Width verifies the referent id width and the per-syntax allocation
// stride: 4 octets / +4 under NDR20, 8 octets / +8 under NDR64.
func TestReferent_Width(t *testing.T) {
	e20 := NewEncoder()
	id0 := e20.nextReferent()
	id1 := e20.nextReferent()
	if id0 != firstReferentID || id1 != firstReferentID+4 {
		t.Errorf("NDR20 referent stride: got %#x,%#x want %#x,%#x", id0, id1, firstReferentID, firstReferentID+4)
	}
	e20b := NewEncoder()
	e20b.writeReferent(firstReferentID)
	if want := []byte{0x00, 0x00, 0x02, 0x00}; !bytes.Equal(e20b.Bytes(), want) {
		t.Errorf("NDR20 referent bytes:\n got %x\nwant %x", e20b.Bytes(), want)
	}

	e64 := NewEncoderForSyntax(NDR64)
	jd0 := e64.nextReferent()
	jd1 := e64.nextReferent()
	if jd0 != firstReferentID || jd1 != firstReferentID+8 {
		t.Errorf("NDR64 referent stride: got %#x,%#x want %#x,%#x", jd0, jd1, firstReferentID, firstReferentID+8)
	}
	e64b := NewEncoderForSyntax(NDR64)
	e64b.writeReferent(firstReferentID)
	if want := []byte{0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00}; !bytes.Equal(e64b.Bytes(), want) {
		t.Errorf("NDR64 referent bytes:\n got %x\nwant %x", e64b.Bytes(), want)
	}
}

// TestWString_NDR64Counts verifies the conformant-varying string framing carries
// 8-octet counts under NDR64 while the UTF-16LE body is unchanged.
func TestWString_NDR64Counts(t *testing.T) {
	e := NewEncoderForSyntax(NDR64)
	e.writeWString("AB")
	want := []byte{
		0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // maximum_count (3, incl NUL)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // offset
		0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // actual_count
		0x41, 0x00, 0x42, 0x00, // "AB"
		0x00, 0x00, // terminator
	}
	if !bytes.Equal(e.Bytes(), want) {
		t.Errorf("NDR64 wstring:\n got %x\nwant %x", e.Bytes(), want)
	}
	got, err := NewDecoderForSyntax(e.Bytes(), NDR64).readWString()
	if err != nil {
		t.Fatalf("NDR64 readWString: %v", err)
	}
	if got != "AB" {
		t.Errorf("NDR64 wstring round trip: got %q want %q", got, "AB")
	}
}
