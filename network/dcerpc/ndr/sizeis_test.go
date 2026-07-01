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

// TestSizeIs_PointerSiblingIndependentBounds verifies that a conformant-varying array whose
// size_is/length_is targets are *pointers* to integers (MS-RRP's [in, out, unique] LPDWORD
// count parameters) derives its maximum_count and actual_count from the pointed-to values,
// independent of the Go slice length. This is the BaseRegQueryValue shape: a full-capacity
// buffer is offered (maximum_count = *lpcbData) while no input octets are transmitted on a
// read (actual_count = *lpcbLen = 0). Before pointer siblings were resolved, both counts
// fell back to len(Data), sending a spurious actual_count and an input body that a DC
// rejects with nca_s_fault_ndr.
func TestSizeIs_PointerSiblingIndependentBounds(t *testing.T) {
	type req struct {
		Data   []uint8 `ndr:"unique,size_is=CbData,varying,length_is=CbLen"`
		CbData *DWORD  `ndr:"unique"`
		CbLen  *DWORD  `ndr:"unique"`
	}
	capacity := DWORD(8)
	valid := DWORD(0)
	// Data carries 8 octets, but only *CbLen (0) are valid; none must be transmitted.
	in := &req{Data: []byte{1, 2, 3, 4, 5, 6, 7, 8}, CbData: &capacity, CbLen: &valid}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // Data referent id
		0x08, 0x00, 0x00, 0x00, // Data maximum_count == *CbData (8), not len(Data)
		0x00, 0x00, 0x00, 0x00, // Data offset
		0x00, 0x00, 0x00, 0x00, // Data actual_count == *CbLen (0); no data body follows
		0x04, 0x00, 0x02, 0x00, // CbData referent id
		0x08, 0x00, 0x00, 0x00, // CbData value
		0x08, 0x00, 0x02, 0x00, // CbLen referent id
		0x00, 0x00, 0x00, 0x00, // CbLen value
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("pointer-sibling size_is/length_is:\n got %x\nwant %x", raw, want)
	}
}
