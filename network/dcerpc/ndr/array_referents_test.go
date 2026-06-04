package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// TestArrayOfUniquePointers verifies that a conformant array of [unique] pointers
// transmits a referent id per element inline (0 for a NULL element) and the referent
// bodies after the whole array, in element order (issue #419, [C706] section 14.3.10).
func TestArrayOfUniquePointers(t *testing.T) {
	type rec struct {
		P []*uint32 `ndr:"conformant"`
	}
	a, b := uint32(0xAA), uint32(0xCC)
	in := &rec{P: []*uint32{&a, nil, &b}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x03, 0, 0, 0, // hoisted maximum_count
		0x00, 0x00, 0x02, 0x00, // P[0] referent id
		0x00, 0x00, 0x00, 0x00, // P[1] NULL
		0x04, 0x00, 0x02, 0x00, // P[2] referent id
		0xAA, 0, 0, 0, // P[0] body (after the whole array)
		0xCC, 0, 0, 0, // P[2] body
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("array of unique pointers:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.P, in.P) {
		t.Errorf("round trip: got %v want %v", out.P, in.P)
	}
}

// TestArrayOfRefPointers verifies that a conformant array of [ref] pointers transmits
// no per-element referent id (the array-of-reference-pointers special case in [C706]
// section 14.3.10), only the referent bodies in element order.
func TestArrayOfRefPointers(t *testing.T) {
	type rec struct {
		P []*uint32 `ndr:"conformant,elem=ref"`
	}
	a, b := uint32(0x11), uint32(0x22)
	in := &rec{P: []*uint32{&a, &b}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0, 0, 0, // hoisted maximum_count
		0x11, 0, 0, 0, // P[0] body — no referent ids transmitted
		0x22, 0, 0, 0, // P[1] body
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("array of ref pointers:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.P, in.P) {
		t.Errorf("round trip: got %v want %v", out.P, in.P)
	}
}

// TestArrayOfStructsWithEmbeddedPointer is the core regression for issue #419: an array
// of structs that each embed a [unique] pointer. NDR requires every element's fixed
// part first, then every element's deferred referent — the referents must not be
// interleaved between element bodies.
func TestArrayOfStructsWithEmbeddedPointer(t *testing.T) {
	type elem struct {
		X uint32
		P *uint32 `ndr:"unique"`
	}
	type rec struct {
		Items []elem `ndr:"conformant"`
	}
	x0, x1 := uint32(0x55), uint32(0x66)
	in := &rec{Items: []elem{{X: 1, P: &x0}, {X: 2, P: &x1}}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0, 0, 0, // hoisted maximum_count
		0x01, 0, 0, 0, // Items[0].X
		0x00, 0x00, 0x02, 0x00, // Items[0].P referent id
		0x02, 0, 0, 0, // Items[1].X
		0x04, 0x00, 0x02, 0x00, // Items[1].P referent id
		0x55, 0, 0, 0, // Items[0].P body (after the whole array, not after Items[0])
		0x66, 0, 0, 0, // Items[1].P body
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("array of structs with embedded pointer:\n got %x\nwant %x", raw, want)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.Items, in.Items) {
		t.Errorf("round trip: got %+v want %+v", out.Items, in.Items)
	}
}

// TestArrayOfStructsWithVaryingString exercises the lsarpc/srvsvc enumeration shape: an
// array of structs that each carry a [unique] pointer to a conformant-varying array
// (here a wchar buffer). All buffers must follow the entire struct array.
func TestArrayOfStructsWithVaryingString(t *testing.T) {
	type entry struct {
		ID   uint32
		Name []uint16 `ndr:"unique,varying"`
	}
	type rec struct {
		Entries []entry `ndr:"conformant"`
	}
	in := &rec{Entries: []entry{
		{ID: 0x0A, Name: []uint16{'h', 'i'}},
		{ID: 0x0B, Name: []uint16{'y', 'o'}},
	}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.Entries, in.Entries) {
		t.Errorf("round trip: got %+v want %+v", out.Entries, in.Entries)
	}
	// The two referent ids must appear inline (within the array body) before either
	// buffer body, confirming the buffers were deferred past the whole array.
	firstBuf := bytes.Index(raw, []byte{'h', 0x00, 'i', 0x00})
	secondRef := bytes.Index(raw, []byte{0x04, 0x00, 0x02, 0x00})
	if secondRef < 0 || firstBuf < 0 || secondRef > firstBuf {
		t.Errorf("buffer body emitted before the array's referent ids: raw=%x", raw)
	}
}
