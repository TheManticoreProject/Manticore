package msnrpc

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestDeltaUnionPointerArmsWire verifies NETLOGON_DELTA_UNION's pointer arms inside the
// NETLOGON_DELTA_ENUM array returned by NetrDatabaseDeltas/Sync/Sync2/Redo: each entry
// carries the arm's referent id inline, and the referents follow the whole array, in
// element order ([C706] 14.3.12.3).
func TestDeltaUnionPointerArmsWire(t *testing.T) {
	entry := func(low uint32) NETLOGON_DELTA_ENUM {
		return NETLOGON_DELTA_ENUM{
			DeltaType: SerialNumberSkip,
			DeltaID:   NETLOGON_DELTA_ID_UNION{Tag: SerialNumberSkip}, // [default] arm: empty
			DeltaUnion: NETLOGON_DELTA_UNION{
				Tag:                   SerialNumberSkip,
				DeltaSerialNumberSkip: &NLPR_MODIFIED_COUNT{ModifiedCount: OLD_LARGE_INTEGER{LowPart: ndr.DWORD(low), HighPart: 1}},
			},
		}
	}
	type param struct{ P NETLOGON_DELTA_ENUM_ARRAY }
	in := &param{P: NETLOGON_DELTA_ENUM_ARRAY{CountReturned: 2, Deltas: []NETLOGON_DELTA_ENUM{entry(0x11111111), entry(0x22222222)}}}
	want := []byte{
		2, 0, 0, 0, // CountReturned
		0x00, 0x00, 0x02, 0x00, // Deltas referent id
		2, 0, 0, 0, // Deltas maximum_count
		// [0]
		22, 0, // DeltaType
		22, 0, // DeltaID discriminant (no arm)
		22, 0, 0, 0, // DeltaUnion discriminant, padded to the pointer arm
		0x04, 0x00, 0x02, 0x00, // DeltaSerialNumberSkip referent id
		// [1]
		22, 0, 22, 0, 22, 0, 0, 0,
		0x08, 0x00, 0x02, 0x00,
		// referents, after every element
		0x11, 0x11, 0x11, 0x11, 1, 0, 0, 0, // [0] NLPR_MODIFIED_COUNT
		0x22, 0x22, 0x22, 0x22, 1, 0, 0, 0, // [1] NLPR_MODIFIED_COUNT
	}
	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("NETLOGON_DELTA_ENUM_ARRAY wire mismatch\n got %x\nwant %x", raw, want)
	}
	var out param
	if err := ndr.Unmarshal(want, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(&out, in) {
		t.Errorf("round trip: got %+v want %+v", out, *in)
	}
}
