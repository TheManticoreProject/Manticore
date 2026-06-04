package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// TestConformantArray_Uint64DeterminantAligned verifies the maximum_count of a
// conformant array of 8-byte elements is placed on an 8-octet boundary (issue #411).
func TestConformantArray_Uint64Determinant(t *testing.T) {
	type rec struct {
		Lead uint8
		P    []uint64 `ndr:"unique,conformant"`
	}
	in := &rec{Lead: 0x01, P: []uint64{0x1122334455667788}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x01, 0x00, 0x00, 0x00, // Lead (uint8) + 3 pad to 4-align the referent id
		0x00, 0x00, 0x02, 0x00, // referent id
		0x01, 0x00, 0x00, 0x00, // maximum_count (4 octets) ...
		0x00, 0x00, 0x00, 0x00, // ... then 4 pad so elements are 8-aligned
		0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, // P[0] (uint64)
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("uint64 conformant array:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Lead != 1 || !reflect.DeepEqual(out.P, in.P) {
		t.Errorf("round trip: got %+v want %+v", out, in)
	}
}

// TestConformantArray_Uint32DeterminantUnchanged confirms <=4-byte elements are
// unaffected (the determinant stays 4-aligned, no extra padding).
func TestConformantArray_Uint32DeterminantUnchanged(t *testing.T) {
	type rec struct {
		Lead uint8
		P    []uint32 `ndr:"unique,conformant"`
	}
	raw, err := Marshal(&rec{Lead: 0x01, P: []uint32{0xAABBCCDD}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x01, 0x00, 0x00, 0x00, // Lead + pad
		0x00, 0x00, 0x02, 0x00, // referent id
		0x01, 0x00, 0x00, 0x00, // maximum_count (no extra pad: elements are 4-aligned)
		0xDD, 0xCC, 0xBB, 0xAA, // P[0]
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("uint32 conformant array:\n got %x\nwant %x", raw, want)
	}
}
