package mswcce

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestCERTTRANSBLOB_RoundTrip exercises the [unique] pointer to a conformant byte
// array sized by the sibling Cb field. Cb must be derived from the slice on
// marshal (the size_is reference points at the exported Go field name Cb, not
// the IDL's lowercase cb).
func TestCERTTRANSBLOB_RoundTrip(t *testing.T) {
	in := CERTTRANSBLOB{Pb: []uint8{0x30, 0x82, 0x01, 0x0a, 0xff}}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out CERTTRANSBLOB
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Cb != ndr.DWORD(len(in.Pb)) {
		t.Errorf("Cb = %d, want %d (count must be derived from the slice)", out.Cb, len(in.Pb))
	}
	if !reflect.DeepEqual(out.Pb, in.Pb) {
		t.Errorf("Pb round-trip: got %v want %v", out.Pb, in.Pb)
	}
}

// TestCERTTRANSBLOB_Empty exercises the null/empty buffer case: a zero-length
// blob must round-trip with Cb == 0.
func TestCERTTRANSBLOB_Empty(t *testing.T) {
	in := CERTTRANSBLOB{}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out CERTTRANSBLOB
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Cb != 0 {
		t.Errorf("Cb = %d, want 0", out.Cb)
	}
	if len(out.Pb) != 0 {
		t.Errorf("Pb len = %d, want 0", len(out.Pb))
	}
}
