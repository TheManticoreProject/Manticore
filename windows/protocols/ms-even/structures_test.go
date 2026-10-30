package mseven

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals v, unmarshals into a fresh value of the same type. v and out must be
// pointers to the same struct type.
func roundTrip(t *testing.T, v any, out any) {
	t.Helper()
	raw, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%T): %v", v, err)
	}
	if err := ndr.Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal(%T): %v", v, err)
	}
}

// TestRPCClientID_RoundTrip exercises RPC_CLIENT_ID ([MS-EVEN] 2.2.6): two scalar DWORDs.
func TestRPCClientID_RoundTrip(t *testing.T) {
	in := RPC_CLIENT_ID{UniqueProcess: 0x11223344, UniqueThread: 0x55667788}
	var out RPC_CLIENT_ID
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("RPC_CLIENT_ID round-trip: got %+v want %+v", out, in)
	}
}

// TestRPCString_RoundTrip exercises RPC_STRING ([MS-EVEN] 2.2.7): a counted ASCII string
// whose Buffer is a [unique] pointer to a conformant char array bounded by MaximumLength,
// including the empty (nil pointer) case.
func TestRPCString_RoundTrip(t *testing.T) {
	cases := []RPC_STRING{
		{Length: 5, MaximumLength: 5, Buffer: []byte("hello")},
		{Length: 1, MaximumLength: 4, Buffer: []byte{0x41}},
		{Length: 0, MaximumLength: 0, Buffer: nil},
	}
	for _, in := range cases {
		var out RPC_STRING
		roundTrip(t, &in, &out)
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("RPC_STRING round-trip: got %+v want %+v", out, in)
		}
	}
}

// TestIELFHandle_RoundTrip exercises IELF_HANDLE ([MS-EVEN]): a 20-byte RPC context handle.
func TestIELFHandle_RoundTrip(t *testing.T) {
	type wrap struct{ H IELF_HANDLE }
	var in wrap
	for i := range in.H {
		in.H[i] = byte(i + 1)
	}
	var out wrap
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("IELF_HANDLE round-trip: got %+v want %+v", out, in)
	}
}

// TestRPCString_WireCounts pins RPC_STRING's [size_is(MaximumLength), length_is(Length)]
// bounds: max_count is MaximumLength and actual_count is Length, so only the first
// Length bytes of a larger buffer are transmitted.
func TestRPCString_WireCounts(t *testing.T) {
	type wrap struct{ S RPC_STRING }
	in := wrap{S: RPC_STRING{Length: 1, MaximumLength: 4, Buffer: []byte{0x41, 0x00, 0x00, 0x00}}}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Length (2) + MaximumLength (2) + referent (4), then max_count, offset, actual_count.
	if len(raw) < 20 {
		t.Fatalf("Marshal produced %d bytes", len(raw))
	}
	maxCount := binary.LittleEndian.Uint32(raw[8:12])
	actualCount := binary.LittleEndian.Uint32(raw[16:20])
	if maxCount != 4 || actualCount != 1 {
		t.Errorf("array header = max %d, actual %d; want max 4, actual 1", maxCount, actualCount)
	}
}
