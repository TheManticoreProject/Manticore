package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// TestEmbeddedConformantArray_Uint32 verifies a hoisted conformant array of a
// non-byte scalar element type (issue #397).
func TestEmbeddedConformantArray_Uint32(t *testing.T) {
	type rec struct {
		N    uint32
		Vals []uint32 `ndr:"conformant"`
	}
	in := &rec{N: 0x11, Vals: []uint32{0xAA, 0xBB}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0, 0, 0, // hoisted maximum_count
		0x11, 0, 0, 0, // N
		0xAA, 0, 0, 0, // Vals[0]
		0xBB, 0, 0, 0, // Vals[1]
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("uint32 array:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.N != in.N || !reflect.DeepEqual(out.Vals, in.Vals) {
		t.Errorf("round trip: got %+v want %+v", out, in)
	}
}

// TestPointerConformantArray_Uint32 verifies a unique pointer to a conformant array
// of uint32 (no hoisting; count travels with the referent).
func TestPointerConformantArray_Uint32(t *testing.T) {
	type rec struct {
		P []uint32 `ndr:"unique,conformant"`
	}
	in := &rec{P: []uint32{0x01, 0x02}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // referent id
		0x02, 0, 0, 0, // maximum_count (in referent)
		0x01, 0, 0, 0, // P[0]
		0x02, 0, 0, 0, // P[1]
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("pointer uint32 array:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.P, in.P) {
		t.Errorf("round trip: got %+v want %+v", out.P, in.P)
	}
}

// TestConformantArray_StructElements verifies a conformant array whose elements are
// structs.
func TestConformantArray_StructElements(t *testing.T) {
	type pair struct {
		A uint16
		B uint16
	}
	type rec struct {
		Items []pair `ndr:"conformant"`
	}
	in := &rec{Items: []pair{{1, 2}, {3, 4}}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0, 0, 0, // hoisted maximum_count
		0x01, 0x00, 0x02, 0x00, // pair{1,2}
		0x03, 0x00, 0x04, 0x00, // pair{3,4}
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("struct array:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.Items, in.Items) {
		t.Errorf("round trip: got %+v want %+v", out.Items, in.Items)
	}
}
