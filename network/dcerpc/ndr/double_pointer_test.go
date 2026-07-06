package ndr

import (
	"bytes"
	"testing"
)

// TestTopLevelDoublePointer verifies a top-level pointer-to-unique-pointer-to-scalar
// (**T) emits two nested referent ids followed by the value, and round-trips. This is
// the [MS-MQMP] CACTransferBuffer OBJECTID**/GUID** shape ([C706] 14.3.10), issue #801.
func TestTopLevelDoublePointer(t *testing.T) {
	type topDouble struct {
		PP **uint32
	}
	v := uint32(0xAABBCCDD)
	p := &v
	raw, err := Marshal(&topDouble{PP: &p})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // outer referent id (in place for a top-level [unique])
		0x04, 0x00, 0x02, 0x00, // inner referent id
		0xDD, 0xCC, 0xBB, 0xAA, // the value
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("top-level **T:\n got %x\nwant %x", raw, want)
	}
	var out topDouble
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.PP == nil || *out.PP == nil || **out.PP != v {
		t.Errorf("round trip: got %v", out.PP)
	}
}

// TestEmbeddedDoublePointer verifies an embedded **T emits an outer referent id inline,
// then defers a construction whose body is the inner referent id followed by the value.
func TestEmbeddedDoublePointer(t *testing.T) {
	type inner struct {
		PP **uint32
	}
	type outer struct {
		I inner
	}
	v := uint32(0xAABBCCDD)
	p := &v
	raw, err := Marshal(&outer{I: inner{PP: &p}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // outer referent id (embedded: placeholder, body deferred)
		0x04, 0x00, 0x02, 0x00, // inner referent id (in the deferred referent body)
		0xDD, 0xCC, 0xBB, 0xAA, // the value
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("embedded **T:\n got %x\nwant %x", raw, want)
	}
	var out outer
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.I.PP == nil || *out.I.PP == nil || **out.I.PP != v {
		t.Errorf("round trip: got %v", out.I.PP)
	}
}

// TestDoublePointerNil verifies a nil inner pointer emits an outer referent id then a
// NULL (0) inner referent, and round-trips back to a nil inner pointer.
func TestDoublePointerNil(t *testing.T) {
	type topDouble struct {
		PP **uint32
	}
	var p *uint32 // nil inner pointer, non-nil outer
	raw, err := Marshal(&topDouble{PP: &p})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // outer referent id
		0x00, 0x00, 0x00, 0x00, // inner NULL referent
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("nil inner **T:\n got %x\nwant %x", raw, want)
	}
	var out topDouble
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.PP == nil || *out.PP != nil {
		t.Errorf("round trip: expected non-nil outer, nil inner; got %v", out.PP)
	}
}

// TestDoublePointerToConformantArray verifies a pointer-to-unique-pointer-to-conformant
// -array (**[]T): outer referent id, inner referent id, then the array's maximum_count
// and elements. This is the [MS-MQMP] [size_is(,n)] TYPE** shape, issue #801.
func TestDoublePointerToConformantArray(t *testing.T) {
	type topArr struct {
		PP **[]uint32 `ndr:"unique"`
	}
	arr := []uint32{0x11111111, 0x22222222}
	pa := &arr
	raw, err := Marshal(&topArr{PP: &pa})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x00, 0x00, 0x02, 0x00, // outer referent id
		0x04, 0x00, 0x02, 0x00, // inner referent id
		0x02, 0x00, 0x00, 0x00, // maximum_count = 2
		0x11, 0x11, 0x11, 0x11,
		0x22, 0x22, 0x22, 0x22,
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("**[]T:\n got %x\nwant %x", raw, want)
	}
	var out topArr
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.PP == nil || *out.PP == nil || len(**out.PP) != 2 ||
		(**out.PP)[0] != arr[0] || (**out.PP)[1] != arr[1] {
		t.Errorf("round trip: got %v", out.PP)
	}
}
