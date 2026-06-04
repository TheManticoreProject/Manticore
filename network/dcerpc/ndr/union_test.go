package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// infoUnion is a declarative union switched on Level: arm 1 is a uint32, arm 2 a
// uint16, and the default arm a uint8.
type infoUnion struct {
	Level uint32 `ndr:"switch"`
	AsU32 uint32 `ndr:"case=1"`
	AsU16 uint16 `ndr:"case=2"`
	Other uint8  `ndr:"default"`
}

func TestUnion_Arm_Golden(t *testing.T) {
	type wrap struct{ U infoUnion }
	cases := []struct {
		name string
		u    infoUnion
		want []byte
	}{
		{"case1", infoUnion{Level: 1, AsU32: 0xAABBCCDD}, []byte{0x01, 0, 0, 0, 0xDD, 0xCC, 0xBB, 0xAA}},
		{"case2", infoUnion{Level: 2, AsU16: 0xBEEF}, []byte{0x02, 0, 0, 0, 0xEF, 0xBE}},
		{"default", infoUnion{Level: 9, Other: 0x7F}, []byte{0x09, 0, 0, 0, 0x7F}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw, err := Marshal(&wrap{U: c.u})
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if !bytes.Equal(raw, c.want) {
				t.Errorf("union:\n got %x\nwant %x", raw, c.want)
			}
			var out wrap
			if err := Unmarshal(raw, &out); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if out.U != c.u {
				t.Errorf("round trip: got %+v want %+v", out.U, c.u)
			}
		})
	}
}

// armBody is a sample pointer-arm payload, mirroring the lsarpc info-class unions whose
// arms are [unique] pointers to structures.
type armBody struct {
	X uint32
}

type policyUnion struct {
	Level uint32   `ndr:"switch"`
	A     *armBody `ndr:"case=1,unique"`
	Name  *WSTR    `ndr:"case=2,unique"`
}

// TestUnion_PointerArm verifies a union whose selected arm is a [unique] pointer to a
// struct: the referent id is inline after the discriminant and the body is deferred to
// the end of the union.
func TestUnion_PointerArm(t *testing.T) {
	type wrap struct {
		U *policyUnion `ndr:"unique"`
	}
	in := &wrap{U: &policyUnion{Level: 1, A: &armBody{X: 0x12345678}}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // U referent id (the pointer to the union)
		0x01, 0, 0, 0, // discriminant Level = 1
		0x04, 0x00, 0x02, 0x00, // arm A referent id
		0x78, 0x56, 0x34, 0x12, // arm A body (X)
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("pointer-arm union:\n got %x\nwant %x", raw, want)
	}
	var out wrap
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.U == nil || out.U.Level != 1 || out.U.A == nil || out.U.A.X != 0x12345678 || out.U.Name != nil {
		t.Errorf("round trip mismatch: %+v", out.U)
	}
}

// TestUnion_Embedded verifies a union embedded as a value field between other fields of
// a structure (the LSA_FOREST_TRUST_RECORD shape): the union must consume exactly its
// discriminant + arm so the trailing field decodes from the right offset.
func TestUnion_Embedded(t *testing.T) {
	type record struct {
		Flags uint32
		Data  infoUnion
		Tail  uint32
	}
	in := &record{Flags: 0x11, Data: infoUnion{Level: 2, AsU16: 0x0203}, Tail: 0x99}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out record
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != *in {
		t.Errorf("round trip: got %+v want %+v", out, *in)
	}
}

// TestUnion_Alignment verifies the union is aligned to the largest alignment of the
// discriminant and any arm (here an 8-octet arm forces 8-alignment of the union body),
// computed over all arms so it does not depend on the selected one.
func TestUnion_Alignment(t *testing.T) {
	type bigUnion struct {
		Sel uint32 `ndr:"switch"`
		Big uint64 `ndr:"case=1"`
	}
	type wrap struct {
		Pre uint32
		U   bigUnion
	}
	in := &wrap{Pre: 0xAA, U: bigUnion{Sel: 1, Big: 0x1122334455667788}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0xAA, 0, 0, 0, // Pre
		0x00, 0, 0, 0, // pad: union aligned to 8 before the discriminant
		0x01, 0, 0, 0, // discriminant Sel = 1
		0x00, 0, 0, 0, // pad: uint64 arm aligned to 8
		0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, // arm Big
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("union alignment:\n got %x\nwant %x", raw, want)
	}
	var out wrap
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out, *in) {
		t.Errorf("round trip: got %+v want %+v", out, *in)
	}
}
