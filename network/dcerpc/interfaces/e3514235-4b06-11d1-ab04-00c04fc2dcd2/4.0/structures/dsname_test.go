package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// TestUUIDWireForm checks UUID holds the 16-octet RPC wire form (guid.GUID.ToBytes
// order) and round-trips back to the same guid.GUID — the encoding the trailing-uint64
// guid.GUID layout would corrupt if used directly under NDR.
func TestUUIDWireForm(t *testing.T) {
	g, err := guid.FromFormatD("e24d201a-4fd6-11d1-a3da-0000f875ae0d")
	if err != nil {
		t.Fatalf("parse guid: %v", err)
	}
	u := UUIDFromGUID(*g)
	if len(u.Octets) != 16 {
		t.Fatalf("UUID is %d octets, want 16", len(u.Octets))
	}
	// Data1 little-endian (e24d201a -> 1a 20 4d e2), Data4 big-endian (a3 da ...).
	want := []byte{0x1a, 0x20, 0x4d, 0xe2, 0xd6, 0x4f, 0xd1, 0x11, 0xa3, 0xda, 0x00, 0x00, 0xf8, 0x75, 0xae, 0x0d}
	for i, b := range want {
		if u.Octets[i] != b {
			t.Fatalf("octet %d = 0x%02x, want 0x%02x (wire order wrong)", i, u.Octets[i], b)
		}
	}
	back := u.GUID()
	if back.ToFormatD() != g.ToFormatD() {
		t.Errorf("round trip: %s != %s", back.ToFormatD(), g.ToFormatD())
	}
}

// TestDSNameGUIDRoundTrip checks a GUID-addressed DSNAME (EXOP_REPL_OBJ form) round-trips
// through NDR behind a unique pointer — exercising the referent id, the hoisted
// maximum_count of the trailing conformant StringName, and the 16-octet GUID.
func TestDSNameGUIDRoundTrip(t *testing.T) {
	type wrapper struct {
		PNC *DSNAME `ndr:"unique"`
	}
	g, _ := guid.FromFormatD("00112233-4455-6677-8899-aabbccddeeff")
	in := NewDSNameFromGUID(*g)
	w := wrapper{PNC: &in}

	raw, err := ndr.Marshal(&w)
	if err != nil {
		t.Fatalf("ndr.Marshal: %v", err)
	}
	var got wrapper
	if err := ndr.Unmarshal(raw, &got); err != nil {
		t.Fatalf("ndr.Unmarshal: %v", err)
	}
	if got.PNC == nil {
		t.Fatal("PNC came back nil")
	}
	if !reflect.DeepEqual(in, *got.PNC) {
		t.Errorf("DSNAME round trip mismatch:\n in = %+v\nout = %+v", in, *got.PNC)
	}
	outGUID := got.PNC.Guid.GUID()
	if outGUID.ToFormatD() != g.ToFormatD() {
		t.Errorf("GUID corrupted: %s != %s", outGUID.ToFormatD(), g.ToFormatD())
	}
}
