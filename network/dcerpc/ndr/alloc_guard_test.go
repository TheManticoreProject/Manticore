package ndr

import (
	"encoding/binary"
	"testing"
)

// TestUnmarshal_HugeArrayCountRejected verifies a conformant array whose wire count
// exceeds the remaining input is rejected instead of triggering an oversized
// allocation (issue #410).
func TestUnmarshal_HugeArrayCountRejected(t *testing.T) {
	type rec struct {
		P []uint32 `ndr:"unique,conformant"`
	}
	// Referent id (non-null) + a maximum_count of 0xFFFFFFFF with no element data.
	stub := make([]byte, 8)
	binary.LittleEndian.PutUint32(stub[0:4], 0x00020000) // referent id
	binary.LittleEndian.PutUint32(stub[4:8], 0xFFFFFFFF) // bogus maximum_count

	var out rec
	if err := Unmarshal(stub, &out); err == nil {
		t.Fatal("Unmarshal with an oversized array count: error = nil, want non-nil")
	}
}

// TestUnmarshal_HugeVaryingCountRejected does the same for a conformant-varying array.
func TestUnmarshal_HugeVaryingCountRejected(t *testing.T) {
	type rec struct {
		P []uint16 `ndr:"unique,varying"`
	}
	stub := make([]byte, 16)
	binary.LittleEndian.PutUint32(stub[0:4], 0x00020000)   // referent id
	binary.LittleEndian.PutUint32(stub[4:8], 0xFFFFFFFF)   // maximum_count
	binary.LittleEndian.PutUint32(stub[8:12], 0)           // offset
	binary.LittleEndian.PutUint32(stub[12:16], 0xFFFFFFFF) // actual_count

	var out rec
	if err := Unmarshal(stub, &out); err == nil {
		t.Fatal("Unmarshal with an oversized varying count: error = nil, want non-nil")
	}
}

// TestUnmarshal_ValidArrayStillWorks confirms a legitimate count is unaffected.
func TestUnmarshal_ValidArrayStillWorks(t *testing.T) {
	type rec struct {
		P []uint32 `ndr:"unique,conformant"`
	}
	in := &rec{P: []uint32{1, 2, 3}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal of a valid array: %v", err)
	}
	if len(out.P) != 3 || out.P[2] != 3 {
		t.Errorf("got %v", out.P)
	}
}
