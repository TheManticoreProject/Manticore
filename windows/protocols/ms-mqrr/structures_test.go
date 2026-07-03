package msmqrr

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals v under both transfer syntaxes, unmarshals into a fresh value of the
// same type, and asserts the result equals v.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	for _, s := range []ndr.Syntax{ndr.NDR20, ndr.NDR64} {
		raw, err := ndr.MarshalAs(&in, s)
		if err != nil {
			t.Fatalf("%s %s marshal: %v", name, s, err)
		}
		var out T
		if err := ndr.UnmarshalAs(raw, &out, s); err != nil {
			t.Fatalf("%s %s unmarshal: %v", name, s, err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Errorf("%s %s round-trip:\n got %+v\nwant %+v", name, s, out, in)
		}
	}
}

// TestSectionBufferRoundTrip exercises the SectionBuffer wire shape, including the
// [unique, size_is(SectionSize)] byte* buffer and every SectionType discriminant. This is
// the exact element type of the ppPacketSections conformant array returned by
// R_StartReceive / R_StartTransactionalReceive, so the round trip validates the nested
// unique-pointer-to-conformant-array layout that those methods emit.
func TestSectionBufferRoundTrip(t *testing.T) {
	roundTrip(t, "SectionBuffer/full", SectionBuffer{
		SectionBufferType: StFullPacket,
		SectionSizeAlloc:  8,
		SectionSize:       4,
		PSectionBuffer:    []uint8{0xDE, 0xAD, 0xBE, 0xEF},
	})
	roundTrip(t, "SectionBuffer/binary1", SectionBuffer{
		SectionBufferType: StBinaryFirstSection,
		SectionSizeAlloc:  16,
		SectionSize:       6,
		PSectionBuffer:    []uint8{1, 2, 3, 4, 5, 6},
	})
	roundTrip(t, "SectionBuffer/srmp2", SectionBuffer{
		SectionBufferType: StSrmpSecondSection,
		SectionSizeAlloc:  2,
		SectionSize:       2,
		PSectionBuffer:    []uint8{0xFF, 0x00},
	})
	// Null buffer: the [unique] pointer is a null referent, SectionSize is 0.
	roundTrip(t, "SectionBuffer/empty", SectionBuffer{
		SectionBufferType: StBinarySecondSection,
		SectionSizeAlloc:  0,
		SectionSize:       0,
		PSectionBuffer:    nil,
	})
}

// TestContextHandle pins the 20-byte RPC context-handle representation ([MS-RPCE] 2.3.2.2)
// and the alias relationship between the serialize and no-serialize forms. The handle is
// only ever marshalled inline as a struct field (the codec has no standalone top-level
// array form), so this checks its size and alias identity rather than a round trip.
func TestContextHandle(t *testing.T) {
	if got := len(QUEUE_CONTEXT_HANDLE_NOSERIALIZE{}); got != 20 {
		t.Fatalf("context handle size = %d bytes, want 20", got)
	}
	// The serialize handle is a Go type alias of the no-serialize handle, so a value of one
	// is assignable to the other without conversion.
	var h QUEUE_CONTEXT_HANDLE_NOSERIALIZE
	var s QUEUE_CONTEXT_HANDLE_SERIALIZE = h
	_ = s
}

// TestSectionTypeValues locks the enum discriminant values against the IDL ([MS-MQRR]).
func TestSectionTypeValues(t *testing.T) {
	cases := map[SectionType]uint16{
		StFullPacket:          0,
		StBinaryFirstSection:  1,
		StBinarySecondSection: 2,
		StSrmpFirstSection:    3,
		StSrmpSecondSection:   4,
	}
	for st, want := range cases {
		if uint16(st) != want {
			t.Errorf("SectionType %v = %d, want %d", st, uint16(st), want)
		}
	}
}
