package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestEFS_RPC_BLOB_RoundTrip exercises a [unique] pointer to a conformant byte
// array sized by a sibling count field. It also asserts that CbData is derived
// from the slice on marshal (the size_is reference must point at the exported
// Go field name CbData, not the IDL's lowercase cbData).
func TestEFS_RPC_BLOB_RoundTrip(t *testing.T) {
	in := EFS_RPC_BLOB{BData: []uint8{0xde, 0xad, 0xbe, 0xef, 0x01}}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out EFS_RPC_BLOB
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.CbData != ndr.DWORD(len(in.BData)) {
		t.Errorf("CbData = %d, want %d (count must be derived from the slice)", out.CbData, len(in.BData))
	}
	if !reflect.DeepEqual(out.BData, in.BData) {
		t.Errorf("BData round-trip: got %v want %v", out.BData, in.BData)
	}
}

// TestENCRYPTION_PROTECTOR_LIST_RoundTrip exercises a count field plus a [unique]
// pointer to a conformant array of [unique] pointer-bearing structs.
func TestENCRYPTION_PROTECTOR_LIST_RoundTrip(t *testing.T) {
	in := ENCRYPTION_PROTECTOR_LIST{
		PProtectors: []*ENCRYPTION_PROTECTOR{
			{CbTotalLength: 12},
			{CbTotalLength: 34},
		},
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out ENCRYPTION_PROTECTOR_LIST
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.NProtectors != 2 || len(out.PProtectors) != 2 {
		t.Fatalf("NProtectors=%d len(PProtectors)=%d, want 2/2", out.NProtectors, len(out.PProtectors))
	}
	if out.PProtectors[0].CbTotalLength != 12 || out.PProtectors[1].CbTotalLength != 34 {
		t.Errorf("element round-trip mismatch: %+v", out.PProtectors)
	}
}
