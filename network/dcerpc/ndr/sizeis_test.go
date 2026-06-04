package ndr

import (
	"bytes"
	"testing"
)

// TestSizeIs_CountDerivedFromArray verifies a size_is target field is written from
// the array length, so the caller need not set it and it cannot disagree (issue #401).
func TestSizeIs_CountDerivedFromArray(t *testing.T) {
	type blob struct {
		CbData uint32 `ndr:"dword"` // size_is target for BData
		BData  []byte `ndr:"unique,conformant,size_is=CbData"`
	}
	// CbData deliberately left 0; it must be derived as len(BData)=3.
	in := &blob{BData: []byte{0xaa, 0xbb, 0xcc}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x03, 0x00, 0x00, 0x00, // CbData, derived from len(BData)
		0x00, 0x00, 0x02, 0x00, // BData referent id
		0x03, 0x00, 0x00, 0x00, // BData maximum_count
		0xaa, 0xbb, 0xcc, // BData elements
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("size_is derivation:\n got %x\nwant %x", raw, want)
	}

	var out blob
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.CbData != 3 || !bytes.Equal(out.BData, in.BData) {
		t.Errorf("round trip: CbData=%d BData=%x", out.CbData, out.BData)
	}
}

// TestSizeIs_LiteralConstant verifies that a literal size_is(<N>) emits the constant N as
// the conformant maximum_count while actual_count reflects the elements actually present,
// matching MS-SAMR's [in, size_is(1000), length_is(Count)] arrays (the server requires the
// fixed bound on the wire). This is the SamrLookupNamesInDomain/SamrLookupIdsInDomain shape.
func TestSizeIs_LiteralConstant(t *testing.T) {
	type req struct {
		Count uint32
		Rids  []uint32 `ndr:"ref,size_is=1000,varying"`
	}
	in := &req{Count: 2, Rids: []uint32{0x1f4, 0x1f5}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0x00, 0x00, 0x00, // Count
		0xe8, 0x03, 0x00, 0x00, // maximum_count == literal 1000, not len
		0x00, 0x00, 0x00, 0x00, // offset
		0x02, 0x00, 0x00, 0x00, // actual_count == len(Rids)
		0xf4, 0x01, 0x00, 0x00, // Rids[0]
		0xf5, 0x01, 0x00, 0x00, // Rids[1]
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("literal size_is:\n got %x\nwant %x", raw, want)
	}
}
