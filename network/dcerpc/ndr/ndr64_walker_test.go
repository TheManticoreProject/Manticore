package ndr

import (
	"bytes"
	"strings"
	"testing"
)

// sidLikeNDR64 mirrors RPC_SID: a structure embedding a conformant array, so its
// maximum_count is hoisted to the front of the struct ([C706] 14.3.3.1) — 8 octets,
// 8-aligned under NDR64.
type sidLikeNDR64 struct {
	Revision uint8
	Count    uint8
	Auth     [6]byte
	Sub      []uint32 `ndr:"conformant,size_is=Count"`
}

// TestNDR64_EmbeddedConformantArray pins the NDR64 layout of a SID-shaped struct: an
// 8-octet hoisted maximum_count, then the fixed members, then the 4-byte elements.
func TestNDR64_EmbeddedConformantArray(t *testing.T) {
	v := sidLikeNDR64{Revision: 1, Auth: [6]byte{0, 0, 0, 0, 0, 5}, Sub: []uint32{32, 544}}
	got, err := MarshalAs(v, NDR64)
	if err != nil {
		t.Fatalf("MarshalAs NDR64: %v", err)
	}
	want := []byte{
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // maximum_count = 2 (8 octets)
		0x01,                               // Revision
		0x02,                               // SubAuthorityCount (derived from len)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x05, // IdentifierAuthority (6 octets, big-endian)
		0x20, 0x00, 0x00, 0x00, // 32
		0x20, 0x02, 0x00, 0x00, // 544
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("NDR64 SID layout:\n got %x\nwant %x", got, want)
	}

	var rt sidLikeNDR64
	if err := UnmarshalAs(got, &rt, NDR64); err != nil {
		t.Fatalf("UnmarshalAs NDR64: %v", err)
	}
	if rt.Revision != 1 || rt.Count != 2 || rt.Auth != v.Auth || len(rt.Sub) != 2 || rt.Sub[0] != 32 || rt.Sub[1] != 544 {
		t.Errorf("NDR64 SID round trip: got %+v want %+v", rt, v)
	}
}

type innerNDR64 struct{ X uint32 }

type outerNDR64 struct {
	A uint32
	P *innerNDR64 `ndr:"unique"`
}

// TestNDR64_UniquePointer pins the NDR64 layout of a [unique] pointer member: a struct
// alignment of 8 (the referent widens to 8 octets), an 8-octet referent id, then the
// pointed-to struct marshalled in place (a top-level pointer's body is not deferred).
func TestNDR64_UniquePointer(t *testing.T) {
	v := outerNDR64{A: 0x11111111, P: &innerNDR64{X: 0x22222222}}
	got, err := MarshalAs(v, NDR64)
	if err != nil {
		t.Fatalf("MarshalAs NDR64: %v", err)
	}
	want := []byte{
		0x11, 0x11, 0x11, 0x11, // A
		0x00, 0x00, 0x00, 0x00, // pad: referent id 8-aligned under NDR64
		0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // referent id (8 octets)
		0x22, 0x22, 0x22, 0x22, // P.X
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("NDR64 unique pointer layout:\n got %x\nwant %x", got, want)
	}

	var rt outerNDR64
	if err := UnmarshalAs(got, &rt, NDR64); err != nil {
		t.Fatalf("UnmarshalAs NDR64: %v", err)
	}
	if rt.A != v.A || rt.P == nil || rt.P.X != v.P.X {
		t.Errorf("NDR64 unique pointer round trip: got %+v want %+v", rt, v)
	}
}

type uniStrNDR64 struct {
	Length uint16
	MaxLen uint16
	Buffer []uint16 `ndr:"unique,varying,size_is=MaxLen/2,length_is=Length/2"`
}

// TestNDR64_ConformantVaryingString round-trips an RPC_UNICODE_STRING-shaped value
// under NDR64, exercising 8-octet referent id, divisor-resolved counts, and the
// conformant-varying framing.
func TestNDR64_ConformantVaryingString(t *testing.T) {
	v := uniStrNDR64{Length: 4, MaxLen: 4, Buffer: []uint16{0x48, 0x69}} // "Hi"
	b, err := MarshalAs(v, NDR64)
	if err != nil {
		t.Fatalf("MarshalAs NDR64: %v", err)
	}
	var rt uniStrNDR64
	if err := UnmarshalAs(b, &rt, NDR64); err != nil {
		t.Fatalf("UnmarshalAs NDR64: %v", err)
	}
	if rt.Length != 4 || rt.MaxLen != 4 || len(rt.Buffer) != 2 || rt.Buffer[0] != 0x48 || rt.Buffer[1] != 0x69 {
		t.Errorf("NDR64 unicode string round trip: got %+v want %+v", rt, v)
	}
}

// TestNDR64_ScalarStructMatchesNDR20 confirms a struct with no counts or referents is
// byte-identical under both syntaxes (nothing widens).
func TestNDR64_ScalarStructMatchesNDR20(t *testing.T) {
	type flat struct {
		A uint16
		B uint32
		C uint64
	}
	v := flat{A: 0xaabb, B: 0xccddeeff, C: 0x0102030405060708}
	n20, err := MarshalAs(v, NDR20)
	if err != nil {
		t.Fatalf("NDR20: %v", err)
	}
	n64, err := MarshalAs(v, NDR64)
	if err != nil {
		t.Fatalf("NDR64: %v", err)
	}
	if !bytes.Equal(n20, n64) {
		t.Errorf("scalar struct differs by syntax:\nNDR20 %x\nNDR64 %x", n20, n64)
	}
}

type unionArm64 struct {
	Tag uint32 `ndr:"switch"`
	A   uint32 `ndr:"case=1"`
}

type hasUnion64 struct {
	U unionArm64
}

type hasPipe64 struct {
	P []byte `ndr:"pipe"`
}

type enumDiscUnion struct {
	Kind uint16 `ndr:"switch,enum"`
	A    uint32 `ndr:"case=1"`
}

type hasEnumUnion struct {
	U enumDiscUnion
}

// TestNDR64_Union verifies a declarative union marshals and round-trips under NDR64.
func TestNDR64_Union(t *testing.T) {
	v := hasUnion64{U: unionArm64{Tag: 1, A: 7}}
	b, err := MarshalAs(v, NDR64)
	if err != nil {
		t.Fatalf("NDR64 union marshal: %v", err)
	}
	var got hasUnion64
	if err := UnmarshalAs(b, &got, NDR64); err != nil {
		t.Fatalf("NDR64 union unmarshal: %v", err)
	}
	if got.U.Tag != 1 || got.U.A != 7 {
		t.Errorf("NDR64 union round trip: got %+v want %+v", got, v)
	}
	if _, err := MarshalAs(v, NDR20); err != nil {
		t.Fatalf("NDR20 union should still marshal: %v", err)
	}
}

// TestNDR64_UnionEnumDiscriminant pins the enum discriminant to 4 octets under NDR64 and
// round-trips. (A 2-octet discriminant is what NDR20 uses; NDR64 widens it.)
func TestNDR64_UnionEnumDiscriminant(t *testing.T) {
	v := hasEnumUnion{U: enumDiscUnion{Kind: 1, A: 0xAABBCCDD}}
	b, err := MarshalAs(v, NDR64)
	if err != nil {
		t.Fatalf("NDR64 marshal: %v", err)
	}
	// disc enum = 4 octets (01 00 00 00), then the uint32 arm (4-aligned).
	want := []byte{0x01, 0x00, 0x00, 0x00, 0xDD, 0xCC, 0xBB, 0xAA}
	if !bytes.Equal(b, want) {
		t.Fatalf("NDR64 enum-discriminant union:\n got  %x\nwant %x", b, want)
	}
	var got hasEnumUnion
	if err := UnmarshalAs(b, &got, NDR64); err != nil {
		t.Fatalf("NDR64 unmarshal: %v", err)
	}
	if got.U.Kind != 1 || got.U.A != 0xAABBCCDD {
		t.Errorf("round trip: got %+v want %+v", got, v)
	}
}

// TestNDR64_PipeRejected verifies a pipe is rejected under NDR64 rather than emitting an
// unverified encoding, while NDR20 still works.
func TestNDR64_PipeRejected(t *testing.T) {
	v := hasPipe64{P: []byte{1, 2, 3}}
	if _, err := MarshalAs(v, NDR20); err != nil {
		t.Fatalf("NDR20 pipe should still marshal: %v", err)
	}
	_, err := MarshalAs(v, NDR64)
	if err == nil || !strings.Contains(err.Error(), "NDR64 pipes") {
		t.Errorf("NDR64 pipe: err = %v, want a 'NDR64 pipes' error", err)
	}
}
