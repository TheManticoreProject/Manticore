package ndr

import (
	"reflect"
	"testing"
)

// itemI64 is a struct whose alignment is 8 (it contains an int64) — like
// PROPERTY_META_DATA_EXT / DS_REPL_CURSOR.
type itemI64 struct {
	V uint32
	T int64
}

// vecI64 mirrors a "vector" struct: a count followed by a conformant array of
// 8-octet-aligned elements (like PROPERTY_META_DATA_EXT_VECTOR).
type vecI64 struct {
	Count uint32
	Items []itemI64 `ndr:"conformant,size_is=Count"`
}

// TestHoistedConformantInt64Alignment guards the fix for the embedded conformant array of
// 8-octet-aligned elements: after the hoisted maximum_count, the structure body (the
// Count member, then the elements) must begin at the element's 8-octet alignment — not
// packed tight after the 4-octet count. Getting this wrong drifts every element (the
// MS-DRSR full-NC metadata bug, #697).
func TestHoistedConformantInt64Alignment(t *testing.T) {
	in := vecI64{Count: 2, Items: []itemI64{{V: 0x11, T: 0x1122334455}, {V: 0x22, T: 0x66778899}}}
	raw, err := Marshal(&in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Expected layout (NDR20): maximum_count@0 (4) + pad(4) + Count@8 (4) + pad(4) +
	// elements@16, each itemI64 = V(4)+pad(4)+T(8) = 16 octets. Total 16 + 2*16 = 48.
	if len(raw) != 48 {
		t.Fatalf("len = %d, want 48 (8-aligned header + 2x16 elements); layout drifted: % x", len(raw), raw)
	}
	var out vecI64
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}
