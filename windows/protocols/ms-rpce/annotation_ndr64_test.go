package msrpce

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestAnnotationWireFormPerSyntax checks the varying-array framing of the
// annotation: 4-octet offset and actual_count under NDR 2.0, 8-octet ones
// 8-aligned under NDR64 ([MS-RPCE] 2.2.5). Reading the NDR64 form with 4-octet
// counts took the high half of the offset as the length and failed every
// ept_lookup made under NDR64.
// annotationOnly wraps an Annotation: the codec marshals structs.
type annotationOnly struct {
	A Annotation
}

func TestAnnotationWireFormPerSyntax(t *testing.T) {
	a := annotationOnly{A: "Fw APIs"}
	for _, tc := range []struct {
		syntax ndr.Syntax
		want   []byte
	}{
		{ndr.NDR20, append([]byte{0, 0, 0, 0, 8, 0, 0, 0}, "Fw APIs\x00"...)},
		{ndr.NDR64, append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0}, "Fw APIs\x00"...)},
	} {
		got, err := ndr.MarshalAs(&a, tc.syntax)
		if err != nil {
			t.Fatalf("syntax %v: MarshalAs: %v", tc.syntax, err)
		}
		if !bytes.Equal(got, tc.want) {
			t.Errorf("syntax %v:\n got %x\nwant %x", tc.syntax, got, tc.want)
		}
		var back annotationOnly
		if err := ndr.UnmarshalAs(tc.want, &back, tc.syntax); err != nil || back != a {
			t.Errorf("syntax %v: UnmarshalAs = %q, %v", tc.syntax, back, err)
		}
	}
}

// TestEptEntryListRoundTripsUnderNDR64 round-trips a lookup-shaped array of
// entries, mixing towers, null towers and annotations of different lengths. It
// is a consistency check of encoder against decoder under both syntaxes; the
// wire form itself is pinned by TestAnnotationWireFormPerSyntax, since a
// symmetric round-trip passes even when both sides use the wrong width.
func TestEptEntryListRoundTripsUnderNDR64(t *testing.T) {
	type list struct {
		Entries []EptEntry `ndr:"varying,inline"`
	}
	in := list{Entries: []EptEntry{
		{Tower: &Twr{TowerLength: 5, TowerOctetString: []byte{1, 0, 0x13, 0, 0}}, Annotation: "Messenger Service"},
		{Annotation: ""},
		{Tower: &Twr{TowerLength: 2, TowerOctetString: []byte{2, 0}}, Annotation: "x"},
	}}
	for _, syntax := range []ndr.Syntax{ndr.NDR20, ndr.NDR64} {
		raw, err := ndr.MarshalAs(&in, syntax)
		if err != nil {
			t.Fatalf("syntax %v: MarshalAs: %v", syntax, err)
		}
		var out list
		if err := ndr.UnmarshalAs(raw, &out, syntax); err != nil {
			t.Fatalf("syntax %v: UnmarshalAs: %v", syntax, err)
		}
		if len(out.Entries) != len(in.Entries) {
			t.Fatalf("syntax %v: %d entries, want %d", syntax, len(out.Entries), len(in.Entries))
		}
		for i := range in.Entries {
			if out.Entries[i].Annotation != in.Entries[i].Annotation || (out.Entries[i].Tower == nil) != (in.Entries[i].Tower == nil) {
				t.Errorf("syntax %v entry %d = %+v, want %+v", syntax, i, out.Entries[i], in.Entries[i])
			}
		}
	}
}

// TestAnnotationRefusesOversizeCount checks the ept_max_annotation_size bound.
func TestAnnotationRefusesOversizeCount(t *testing.T) {
	raw := []byte{0, 0, 0, 0, 0xFF, 0, 0, 0}
	raw = append(raw, bytes.Repeat([]byte{'a'}, 0xFF)...)
	var a annotationOnly
	err := ndr.UnmarshalAs(raw, &a, ndr.NDR20)
	if err == nil || !strings.Contains(err.Error(), "exceed") {
		t.Errorf("an annotation longer than ept_max_annotation_size: err = %v", err)
	}
}
