package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// Regression tests for the deferral scope of a union's pointer arm. A union embedded in a
// structure or an array is part of that construction, so the referent of its pointer arm
// is deferred to the end of the enclosing construction ([C706] section 14.3.12.3), after
// the structure's later fields or after every element of the array — exactly as the
// referent of a pointer inside an embedded structure is.

// unionArmStruct is the structure counterpart of policyUnion selecting arm 1: the same
// inline representation (a uint32 followed by a [unique] pointer to armBody).
type unionArmStruct struct {
	Level uint32
	A     *armBody `ndr:"unique"`
}

// TestUnion_PointerArmDeferredPastEnclosingStruct marshals a union with a pointer arm
// embedded in a structure that has a trailing field. The arm's referent follows the
// trailing field, matching the encoding of the equivalent embedded structure.
func TestUnion_PointerArmDeferredPastEnclosingStruct(t *testing.T) {
	type outerU struct {
		U     policyUnion
		Trail uint32
	}
	type paramU struct{ P outerU }
	type outerS struct {
		S     unionArmStruct
		Trail uint32
	}
	type paramS struct{ P outerS }

	in := &paramU{P: outerU{U: policyUnion{Level: 1, A: &armBody{X: 0x12345678}}, Trail: 0xAABBCCDD}}
	for _, syntax := range []Syntax{NDR20, NDR64} {
		raw, err := MarshalAs(in, syntax)
		if err != nil {
			t.Fatalf("%v: Marshal: %v", syntax, err)
		}
		ref, err := MarshalAs(&paramS{P: outerS{S: unionArmStruct{Level: 1, A: &armBody{X: 0x12345678}}, Trail: 0xAABBCCDD}}, syntax)
		if err != nil {
			t.Fatalf("%v: Marshal (struct): %v", syntax, err)
		}
		if !bytes.Equal(raw, ref) {
			t.Errorf("%v: embedded union pointer arm:\n got %x\nwant %x (equivalent embedded struct)", syntax, raw, ref)
		}
		if syntax == NDR20 {
			want := []byte{
				0x01, 0, 0, 0, // discriminant Level = 1
				0x00, 0x00, 0x02, 0x00, // arm A referent id
				0xDD, 0xCC, 0xBB, 0xAA, // Trail
				0x78, 0x56, 0x34, 0x12, // arm A referent (deferred past Trail)
			}
			if !bytes.Equal(raw, want) {
				t.Errorf("NDR20 golden:\n got %x\nwant %x", raw, want)
			}
		}
		var out paramU
		if err := UnmarshalAs(raw, &out, syntax); err != nil {
			t.Fatalf("%v: Unmarshal: %v", syntax, err)
		}
		if !reflect.DeepEqual(&out, in) {
			t.Errorf("%v: round trip: got %+v want %+v", syntax, out, *in)
		}
	}
}

// TestUnion_PointerArmDeferredPastArrayElements marshals an array of unions with pointer
// arms. Every element's inline part comes first and the arm referents follow the whole
// array, in element order.
func TestUnion_PointerArmDeferredPastArrayElements(t *testing.T) {
	type outerU struct{ Arr []policyUnion }
	type paramU struct{ P outerU }
	type outerS struct{ Arr []unionArmStruct }
	type paramS struct{ P outerS }

	in := &paramU{P: outerU{Arr: []policyUnion{
		{Level: 1, A: &armBody{X: 0x11111111}},
		{Level: 1, A: &armBody{X: 0x22222222}},
	}}}
	for _, syntax := range []Syntax{NDR20, NDR64} {
		raw, err := MarshalAs(in, syntax)
		if err != nil {
			t.Fatalf("%v: Marshal: %v", syntax, err)
		}
		ref, err := MarshalAs(&paramS{P: outerS{Arr: []unionArmStruct{
			{Level: 1, A: &armBody{X: 0x11111111}},
			{Level: 1, A: &armBody{X: 0x22222222}},
		}}}, syntax)
		if err != nil {
			t.Fatalf("%v: Marshal (struct): %v", syntax, err)
		}
		if !bytes.Equal(raw, ref) {
			t.Errorf("%v: array of pointer-arm unions:\n got %x\nwant %x (equivalent struct array)", syntax, raw, ref)
		}
		if syntax == NDR20 {
			want := []byte{
				0x02, 0, 0, 0, // maximum_count
				0x01, 0, 0, 0, // [0] discriminant
				0x00, 0x00, 0x02, 0x00, // [0] arm referent id
				0x01, 0, 0, 0, // [1] discriminant
				0x04, 0x00, 0x02, 0x00, // [1] arm referent id
				0x11, 0x11, 0x11, 0x11, // [0] arm referent
				0x22, 0x22, 0x22, 0x22, // [1] arm referent
			}
			if !bytes.Equal(raw, want) {
				t.Errorf("NDR20 golden:\n got %x\nwant %x", raw, want)
			}
		}
		var out paramU
		if err := UnmarshalAs(raw, &out, syntax); err != nil {
			t.Fatalf("%v: Unmarshal: %v", syntax, err)
		}
		if !reflect.DeepEqual(&out, in) {
			t.Errorf("%v: round trip: got %+v want %+v", syntax, out, *in)
		}
	}
}
