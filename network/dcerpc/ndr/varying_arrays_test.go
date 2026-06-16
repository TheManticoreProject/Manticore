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

// TestInlineConformantVaryingArray_PinsRealWire guards the "inline" tag against the
// circular-test blind spot that hid the EPM ept_lookup/ept_map decode bug (#631, #633).
//
// The blind spot: a test that builds its fixture by marshalling through the codec and then
// unmarshalling with the SAME tag agrees with itself no matter whether the tag matches the
// real wire — so it cannot catch a bare top-level array wrongly modelled as a pointer.
// This test sidesteps that by decoding hand-built octets laid out exactly as a server
// sends them (the conformance count inline, no referent id, not hoisted) and then asserting
// that the *pointer* tagging — the historical bug — cannot read those same octets. A pure
// round-trip could not tell the two tags apart; these ground-truth bytes can.
func TestInlineConformantVaryingArray_PinsRealWire(t *testing.T) {
	// A bare, non-pointer, top-level conformant-varying array parameter: count words inline,
	// no referent id, not hoisted to the struct front. Mirrors ept_lookup's
	// `[out, size_is(max_ents), length_is(*num_ents)] ept_entry_t entries[]`.
	type correct struct {
		Count   uint32
		Entries []uint32 `ndr:"varying,inline"`
		Status  uint32
	}
	// The historical defect: the same array modelled as a pointer-prefixed array.
	type buggy struct {
		Count   uint32
		Entries []uint32 `ndr:"ptr,varying"`
		Status  uint32
	}

	// Ground-truth wire bytes, authored by hand rather than by the encoder:
	//   Count | max_count | offset | actual_count | Entries[0] | Entries[1] | Status
	// max_count (5) is deliberately larger than actual_count (2) — as a real server echoing
	// its size_is cap does — so the test also pins that the decoder sizes the result from
	// actual_count, not max_count.
	wire := []byte{
		0x02, 0, 0, 0, // Count = 2
		0x05, 0, 0, 0, // max_count = 5 (inline conformance, NOT a referent id)
		0x00, 0, 0, 0, // offset = 0
		0x02, 0, 0, 0, // actual_count = 2
		0xAA, 0, 0, 0, // Entries[0]
		0xBB, 0, 0, 0, // Entries[1]
		0x0D, 0xA0, 0xC9, 0x16, // Status = 0x16C9A00D
	}

	var ok correct
	if err := Unmarshal(wire, &ok); err != nil {
		t.Fatalf("inline decode of the real wire layout failed: %v", err)
	}
	if ok.Count != 2 {
		t.Errorf("Count = %d, want 2 (max_count must NOT be hoisted ahead of Count)", ok.Count)
	}
	if ok.Status != 0x16C9A00D {
		t.Errorf("Status = 0x%08x, want 0x16C9A00D", ok.Status)
	}
	if !reflect.DeepEqual(ok.Entries, []uint32{0xAA, 0xBB}) {
		t.Errorf("Entries = %v, want [0xAA 0xBB] (sized from actual_count, not max_count)", ok.Entries)
	}

	// The blind-spot guard: the pointer tagging consumes max_count(5) as a referent id and
	// then misreads the count words, so it CANNOT reproduce the correct decode of the real
	// wire. If it could, the bytes would not discriminate the two tags and no round-trip
	// test could ever have caught the bug.
	var bad buggy
	if err := Unmarshal(wire, &bad); err == nil && reflect.DeepEqual(bad.Entries, []uint32{0xAA, 0xBB}) && bad.Count == 2 && bad.Status == 0x16C9A00D {
		t.Fatal("pointer tagging decoded the inline wire layout correctly; the wire bytes do " +
			"not distinguish ptr from inline, so a codec round-trip could not catch this bug")
	}

	// The encoder must be symmetric with the inline layout: no referent id, the conformance
	// count emitted in place. (Re-encoding writes max_count = len, hence 2 here.)
	got, err := Marshal(&ok)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0, 0, 0, // Count
		0x02, 0, 0, 0, // max_count = len (inline, no referent id)
		0x00, 0, 0, 0, // offset
		0x02, 0, 0, 0, // actual_count
		0xAA, 0, 0, 0, // Entries[0]
		0xBB, 0, 0, 0, // Entries[1]
		0x0D, 0xA0, 0xC9, 0x16, // Status
	}
	if !bytes.Equal(got, want) {
		t.Errorf("inline encode:\n got %x\nwant %x", got, want)
	}
}
