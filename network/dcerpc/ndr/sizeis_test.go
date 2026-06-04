package ndr

import (
	"bytes"
	"testing"
)

// TestSizeIs_CountDerivedFromArray verifies a size_is target field is written from
// the array length, so the caller need not set it and it cannot disagree (issue #401).
func TestSizeIs_CountDerivedFromArray(t *testing.T) {
	type blob struct {
		CbData uint32 `ndr:"dword"` // size_is target for BData
		BData  []byte `ndr:"unique,conformant,size_is=CbData"`
	}
	// CbData deliberately left 0; it must be derived as len(BData)=3.
	in := &blob{BData: []byte{0xaa, 0xbb, 0xcc}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x03, 0x00, 0x00, 0x00, // CbData, derived from len(BData)
		0x00, 0x00, 0x02, 0x00, // BData referent id
		0x03, 0x00, 0x00, 0x00, // BData maximum_count
		0xaa, 0xbb, 0xcc, // BData elements
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("size_is derivation:\n got %x\nwant %x", raw, want)
	}

	var out blob
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.CbData != 3 || !bytes.Equal(out.BData, in.BData) {
		t.Errorf("round trip: CbData=%d BData=%x", out.CbData, out.BData)
	}
}
