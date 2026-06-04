package ndr

import (
	"bytes"
	"testing"
)

// TestTopLevelRefPointer verifies a top-level [ref] pointer is marshalled in place
// with no referent id (issue #398).
func TestTopLevelRefPointer(t *testing.T) {
	type topRef struct {
		P *uint32 `ndr:"ref"`
	}
	v := uint32(0xAABBCCDD)
	raw, err := Marshal(&topRef{P: &v})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{0xDD, 0xCC, 0xBB, 0xAA} // referent in place, no placeholder
	if !bytes.Equal(raw, want) {
		t.Errorf("top-level ref:\n got %x\nwant %x", raw, want)
	}
	var out topRef
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.P == nil || *out.P != v {
		t.Errorf("round trip: got %v", out.P)
	}
}

// TestEmbeddedRefPointer verifies an embedded [ref] pointer emits a 4-octet
// placeholder with its referent deferred (issue #398).
func TestEmbeddedRefPointer(t *testing.T) {
	type inner struct {
		P *uint32 `ndr:"ref"`
	}
	type outer struct {
		I inner
	}
	v := uint32(0xAABBCCDD)
	raw, err := Marshal(&outer{I: inner{P: &v}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // embedded [ref] placeholder (referent id)
		0xDD, 0xCC, 0xBB, 0xAA, // deferred referent
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("embedded ref:\n got %x\nwant %x", raw, want)
	}
	var out outer
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.I.P == nil || *out.I.P != v {
		t.Errorf("round trip: got %v", out.I.P)
	}
}
