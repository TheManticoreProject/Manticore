package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// Regression tests for the deferral scope of pointers inside an array that is embedded in
// a structure. The array is part of the enclosing construction, so the referents of its
// elements' pointers are deferred to the end of that construction ([C706] section
// 14.3.12.3): after the structure's later fields, and after the referents of pointers in
// earlier fields — exactly as for an embedded structure or union.

// ptrElem is an array element carrying a [unique] pointer.
type ptrElem struct {
	X *uint32 `ndr:"unique"`
}

// TestFixedArray_ElementReferentsDeferredPastEnclosingStruct marshals a fixed array of
// pointer-bearing structures followed by a trailing field. The element referents follow
// the trailing field, matching the encoding of the equivalent embedded structure.
func TestFixedArray_ElementReferentsDeferredPastEnclosingStruct(t *testing.T) {
	type outerA struct {
		A [1]ptrElem
		T uint32
	}
	type paramA struct{ P outerA }
	type outerN struct { // nested fixed array: [1][1]
		A [1][1]ptrElem
		T uint32
	}
	type paramN struct{ P outerN }
	type outerS struct {
		A ptrElem
		T uint32
	}
	type paramS struct{ P outerS }

	v := uint32(0x12345678)
	inA := &paramA{P: outerA{A: [1]ptrElem{{X: &v}}, T: 0xAABBCCDD}}
	inN := &paramN{P: outerN{A: [1][1]ptrElem{{{X: &v}}}, T: 0xAABBCCDD}}
	for _, syntax := range []Syntax{NDR20, NDR64} {
		ref, err := MarshalAs(&paramS{P: outerS{A: ptrElem{X: &v}, T: 0xAABBCCDD}}, syntax)
		if err != nil {
			t.Fatalf("%v: Marshal (struct): %v", syntax, err)
		}
		if syntax == NDR20 {
			want := []byte{
				0x00, 0x00, 0x02, 0x00, // A[0].X referent id
				0xDD, 0xCC, 0xBB, 0xAA, // T
				0x78, 0x56, 0x34, 0x12, // A[0].X referent (deferred past T)
			}
			if !bytes.Equal(ref, want) {
				t.Fatalf("NDR20 golden:\n got %x\nwant %x", ref, want)
			}
		}
		for _, tc := range []struct {
			name string
			in   any
			out  any
		}{
			{"[1]", inA, &paramA{}},
			{"[1][1]", inN, &paramN{}},
		} {
			raw, err := MarshalAs(tc.in, syntax)
			if err != nil {
				t.Fatalf("%v %s: Marshal: %v", syntax, tc.name, err)
			}
			if !bytes.Equal(raw, ref) {
				t.Errorf("%v %s: embedded fixed array:\n got %x\nwant %x (equivalent embedded struct)", syntax, tc.name, raw, ref)
			}
			if err := UnmarshalAs(ref, tc.out, syntax); err != nil {
				t.Fatalf("%v %s: Unmarshal: %v", syntax, tc.name, err)
			}
			if !reflect.DeepEqual(tc.out, tc.in) {
				t.Errorf("%v %s: decode: got %+v want %+v", syntax, tc.name, tc.out, tc.in)
			}
		}
	}
}

// TestConformantArray_ElementReferentsFollowEarlierFieldReferents marshals a structure
// with a [unique] pointer field followed by a hoisted conformant array of pointer-bearing
// structures. Referents are deferred in field order, so the earlier field's referent
// precedes the array elements' referents.
func TestConformantArray_ElementReferentsFollowEarlierFieldReferents(t *testing.T) {
	type inner struct {
		P   *uint32 `ndr:"unique"`
		N   uint32
		Arr []ptrElem `ndr:"conformant,size_is=N"`
	}
	type param struct {
		In *inner `ndr:"unique"`
	}
	p, a, b := uint32(0x11111111), uint32(0x22222222), uint32(0x33333333)
	in := &param{In: &inner{P: &p, N: 2, Arr: []ptrElem{{X: &a}, {X: &b}}}}
	for _, syntax := range []Syntax{NDR20, NDR64} {
		raw, err := MarshalAs(in, syntax)
		if err != nil {
			t.Fatalf("%v: Marshal: %v", syntax, err)
		}
		if syntax == NDR20 {
			want := []byte{
				0x00, 0x00, 0x02, 0x00, // In referent id
				0x02, 0, 0, 0, // hoisted maximum_count
				0x04, 0x00, 0x02, 0x00, // P referent id
				0x02, 0, 0, 0, // N
				0x08, 0x00, 0x02, 0x00, // Arr[0].X referent id
				0x0C, 0x00, 0x02, 0x00, // Arr[1].X referent id
				0x11, 0x11, 0x11, 0x11, // P referent (earlier field first)
				0x22, 0x22, 0x22, 0x22, // Arr[0].X referent
				0x33, 0x33, 0x33, 0x33, // Arr[1].X referent
			}
			if !bytes.Equal(raw, want) {
				t.Errorf("NDR20 golden:\n got %x\nwant %x", raw, want)
			}
		}
		var out param
		if err := UnmarshalAs(raw, &out, syntax); err != nil {
			t.Fatalf("%v: Unmarshal: %v", syntax, err)
		}
		if !reflect.DeepEqual(&out, in) {
			t.Errorf("%v: round trip: got %+v want %+v", syntax, out.In, in.In)
		}
	}
}

// TestConformantArray_ChunkedDecodeWithEnclosingQueue decodes a hoisted conformant array
// large enough (in memory) to take unmarshalElements' chunked path, with the element
// referents read from the enclosing structure's queue after the array has been placed.
// Every element's referent must land in the decoded slice.
func TestConformantArray_ChunkedDecodeWithEnclosingQueue(t *testing.T) {
	type bigElem struct {
		Pad [64 << 10]byte `ndr:"skip"` // in-memory weight only, forces chunking
		X   *uint32        `ndr:"unique"`
	}
	type inner struct {
		P   *uint32 `ndr:"unique"`
		N   uint32
		Arr []bigElem `ndr:"conformant,size_is=N"`
	}
	type param struct {
		In *inner `ndr:"unique"`
	}
	const n = 40
	p := uint32(0xFEEDFACE)
	in := &param{In: &inner{P: &p, N: n, Arr: make([]bigElem, n)}}
	for i := range in.In.Arr {
		v := uint32(i + 1)
		in.In.Arr[i].X = &v
	}
	for _, syntax := range []Syntax{NDR20, NDR64} {
		raw, err := MarshalAs(in, syntax)
		if err != nil {
			t.Fatalf("%v: Marshal: %v", syntax, err)
		}
		var out param
		if err := UnmarshalAs(raw, &out, syntax); err != nil {
			t.Fatalf("%v: Unmarshal: %v", syntax, err)
		}
		if out.In == nil || out.In.P == nil || *out.In.P != p || len(out.In.Arr) != n {
			t.Fatalf("%v: decode: got %+v", syntax, out.In)
		}
		for i, e := range out.In.Arr {
			if e.X == nil || *e.X != uint32(i+1) {
				t.Fatalf("%v: Arr[%d].X = %v, want %d", syntax, i, e.X, i+1)
			}
		}
	}
}
