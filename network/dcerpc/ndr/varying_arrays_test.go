package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// TestPointerConformantVaryingArray verifies a unique pointer to a conformant-varying
// array carries maximum_count + offset + actual_count before the elements (issue #396).
func TestPointerConformantVaryingArray(t *testing.T) {
	type rec struct {
		P []uint32 `ndr:"unique,varying"`
	}
	in := &rec{P: []uint32{0xAA, 0xBB}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // referent id
		0x02, 0, 0, 0, // maximum_count
		0x00, 0, 0, 0, // offset
		0x02, 0, 0, 0, // actual_count
		0xAA, 0, 0, 0, // P[0]
		0xBB, 0, 0, 0, // P[1]
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("conformant-varying array:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.P, in.P) {
		t.Errorf("round trip: got %v want %v", out.P, in.P)
	}
}

// TestEmbeddedConformantVaryingArray verifies the hoisted max_count plus in-place
// offset/actual_count framing for an embedded conformant-varying array.
func TestEmbeddedConformantVaryingArray(t *testing.T) {
	type rec struct {
		N    uint32
		Data []uint16 `ndr:"varying"`
	}
	in := &rec{N: 0x09, Data: []uint16{0x01, 0x02, 0x03}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x03, 0, 0, 0, // hoisted maximum_count
		0x09, 0, 0, 0, // N
		0x00, 0, 0, 0, // offset
		0x03, 0, 0, 0, // actual_count
		0x01, 0x00, 0x02, 0x00, 0x03, 0x00, // elements (uint16)
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("embedded conformant-varying:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.N != in.N || !reflect.DeepEqual(out.Data, in.Data) {
		t.Errorf("round trip: got %+v want %+v", out, in)
	}
}
