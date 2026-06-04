package ndr

import (
	"bytes"
	"testing"
)

// TestRetval_EncodedAfterDeferredReferents verifies a field tagged `retval` (the RPC
// return value) is placed after the construction's deferred pointer referents, not
// inline with the other fields.
func TestRetval_EncodedAfterDeferredReferents(t *testing.T) {
	type resp struct {
		Ptr    *uint32 `ndr:"unique"`
		Status uint32  `ndr:"retval"`
	}
	v := uint32(0xAABBCCDD)
	raw, err := Marshal(&resp{Ptr: &v, Status: 0x11})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // Ptr referent id (inline)
		0xDD, 0xCC, 0xBB, 0xAA, // deferred *uint32 referent
		0x11, 0x00, 0x00, 0x00, // retval, after the deferred referent
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("retval ordering:\n got %x\nwant %x", raw, want)
	}

	var out resp
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Ptr == nil || *out.Ptr != v || out.Status != 0x11 {
		t.Errorf("round trip: ptr=%v status=%#x", out.Ptr, out.Status)
	}
}
