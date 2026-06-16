package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// secDescLike mirrors the tag layout of MS-RRP's RPC_SECURITY_DESCRIPTOR (and MS-LSAD's
// LSAPR_CR_CIPHER_VALUE): a [unique] pointer to a conformant-varying byte array whose
// maximum_count (size_is) and actual_count (length_is) name *distinct* sibling fields, so
// the buffer capacity (CbIn) and the valid length (CbOut) are independent values.
type secDescLike struct {
	Buf   []byte `ndr:"unique,size_is=CbIn,varying,length_is=CbOut"`
	CbIn  uint32
	CbOut uint32
}

// secDescCall embeds the descriptor as a field, the way the real BaseRegGetKeySecurity /
// BaseRegSetKeySecurity request structs do. This drives the embedded (deferred-referent)
// marshalling path: the descriptor's inline fields (referent id, CbIn, CbOut) are written
// first, then the array body (maximum_count, offset, actual_count, elements) is deferred
// to the end of the enclosing construction.
type secDescCall struct {
	SD secDescLike
}

// TestIndependentBounds_DistinctSizeAndLength is the regression test for issue #618: a
// conformant-varying byte array whose size_is and length_is reference different fields
// must marshal maximum_count from size_is (CbIn, the capacity) and actual_count from
// length_is (CbOut, the valid length), transmitting only the valid bytes — instead of
// forcing both counts (and both fields) to the slice length.
func TestIndependentBounds_DistinctSizeAndLength(t *testing.T) {
	const refID = "\x00\x00\x02\x00" // firstReferentID, little-endian

	cases := []struct {
		name string
		in   secDescLike
		want []byte
	}{
		{
			// SET shape: capacity == valid length == buffer length.
			name: "full buffer (CbIn==CbOut==len)",
			in:   secDescLike{Buf: []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02}, CbIn: 6, CbOut: 6},
			want: bytes.Join([][]byte{
				[]byte(refID),
				{0x06, 0, 0, 0},                      // CbIn field = 6
				{0x06, 0, 0, 0},                      // CbOut field = 6
				{0x06, 0, 0, 0},                      // maximum_count = CbIn = 6
				{0x00, 0, 0, 0},                      // offset = 0
				{0x06, 0, 0, 0},                      // actual_count = CbOut = 6
				{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02}, // 6 bytes transmitted
			}, nil),
		},
		{
			// GET in-buffer shape: a non-NULL 16-byte capacity buffer with 0 valid bytes.
			// maximum_count advertises the capacity; nothing is transmitted.
			name: "capacity buffer, zero valid (CbIn=16, CbOut=0)",
			in:   secDescLike{Buf: make([]byte, 16), CbIn: 16, CbOut: 0},
			want: bytes.Join([][]byte{
				[]byte(refID),
				{0x10, 0, 0, 0}, // CbIn field = 16
				{0x00, 0, 0, 0}, // CbOut field = 0
				{0x10, 0, 0, 0}, // maximum_count = CbIn = 16 (not the slice length)
				{0x00, 0, 0, 0}, // offset = 0
				{0x00, 0, 0, 0}, // actual_count = CbOut = 0 (no bytes follow)
			}, nil),
		},
		{
			// NULL buffer must still carry the caller's CbIn (the bug overwrote it to 0,
			// which is why the live server reported the wrong required size).
			name: "NULL buffer preserves CbIn (CbIn=4096)",
			in:   secDescLike{Buf: nil, CbIn: 4096, CbOut: 0},
			want: bytes.Join([][]byte{
				{0x00, 0, 0, 0},    // NULL referent id
				{0x00, 0x10, 0, 0}, // CbIn field = 4096
				{0x00, 0, 0, 0},    // CbOut field = 0
			}, nil),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Marshal(&secDescCall{SD: c.in})
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if !bytes.Equal(got, c.want) {
				t.Errorf("wire mismatch:\n got  %x\n want %x", got, c.want)
			}
		})
	}
}

// TestIndependentBounds_RoundTrip confirms the [out] decode path reconstructs the struct:
// the server returns a descriptor whose valid length is shorter than its advertised
// capacity, and only the valid bytes are present on the wire.
func TestIndependentBounds_RoundTrip(t *testing.T) {
	in := secDescCall{SD: secDescLike{Buf: []byte{0x01, 0x02, 0x03}, CbIn: 64, CbOut: 3}}
	raw, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out secDescCall
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// CbOut and the transmitted bytes round-trip; CbIn (capacity) is preserved verbatim.
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip mismatch:\n in:  %+v\n out: %+v", in, out)
	}
}
