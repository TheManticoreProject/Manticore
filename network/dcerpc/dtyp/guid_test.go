package dtyp

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

type guidWrap struct{ G GUID }

// TestGUID_WireFormatFromGuidGUID verifies that a windows/guid.GUID generates the dtyp
// NDR wire format: NewGUID(g) marshals to exactly 16 octets — identical to g.ToBytes()
// — under both NDR20 and NDR64, and round-trips back to the original GUID.
func TestGUID_WireFormatFromGuidGUID(t *testing.T) {
	g, err := guid.FromString("a6500248-1cc1-4e89-b2bf-ebe99677d084")
	if err != nil {
		t.Fatalf("parse guid: %v", err)
	}
	want := g.ToBytes() // canonical 16-octet uuid_t wire form
	if len(want) != 16 {
		t.Fatalf("guid.ToBytes() = %d octets, want 16", len(want))
	}

	d := NewGUID(*g)
	for _, s := range []ndr.Syntax{ndr.NDR20, ndr.NDR64} {
		got, err := ndr.MarshalAs(&guidWrap{G: d}, s)
		if err != nil {
			t.Fatalf("%s marshal: %v", s, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s GUID wire form:\n got  %x\nwant %x", s, got, want)
		}
	}

	// Round-trip: dtyp.GUID back to windows/guid.GUID.
	if back := d.GUID(); !bytes.Equal(back.ToBytes(), want) {
		t.Errorf("round trip: got %x want %x", back.ToBytes(), want)
	}
	if d.String() != "a6500248-1cc1-4e89-b2bf-ebe99677d084" {
		t.Errorf("String() = %q", d.String())
	}
}
