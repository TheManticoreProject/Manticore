package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// TestPipe_ByteWireFormat verifies an NDR byte pipe ([C706] 14.7): a uint32 chunk
// count, the elements, then a terminating 0-count chunk; and that it round-trips.
func TestPipe_ByteWireFormat(t *testing.T) {
	type req struct {
		Handle [4]byte
		Data   []byte `ndr:"pipe"`
	}
	in := &req{Handle: [4]byte{0xaa, 0xbb, 0xcc, 0xdd}, Data: []byte{0x01, 0x02, 0x03}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0xaa, 0xbb, 0xcc, 0xdd, // Handle
		0x03, 0x00, 0x00, 0x00, // chunk count = 3
		0x01, 0x02, 0x03, // elements
		0x00,                   // pad: the next uint32 count aligns to 4 ([C706] 14)
		0x00, 0x00, 0x00, 0x00, // terminating empty chunk (count 0)
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("pipe wire:\n got %x\nwant %x", raw, want)
	}
	var out req
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(out.Data, in.Data) || out.Handle != in.Handle {
		t.Errorf("round trip: got %+v want %+v", out, in)
	}
}

// TestPipe_Empty verifies an empty pipe is a single 0-count chunk and round-trips.
func TestPipe_Empty(t *testing.T) {
	type req struct {
		Data []byte `ndr:"pipe"`
	}
	raw, err := Marshal(&req{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(raw, []byte{0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("empty pipe = %x, want 00000000", raw)
	}
	var out req
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Data) != 0 {
		t.Errorf("empty pipe round trip: got %v", out.Data)
	}
}

// TestPipe_MultiChunk verifies that several chunks concatenate on read.
func TestPipe_MultiChunk(t *testing.T) {
	// Two 4-byte chunks (so each uint32 count stays 4-aligned), then the terminator.
	raw := []byte{
		0x04, 0x00, 0x00, 0x00, 0x11, 0x22, 0x33, 0x44,
		0x04, 0x00, 0x00, 0x00, 0x55, 0x66, 0x77, 0x88,
		0x00, 0x00, 0x00, 0x00,
	}
	type req struct {
		Data []byte `ndr:"pipe"`
	}
	var out req
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(out.Data, []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}) {
		t.Errorf("multi-chunk concat = %x, want 1122334455667788", out.Data)
	}
}

// TestPipe_TypedElements verifies a pipe of a non-byte scalar.
func TestPipe_TypedElements(t *testing.T) {
	type req struct {
		Vals []uint32 `ndr:"pipe"`
	}
	in := &req{Vals: []uint32{0x0a, 0x0b}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out req
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.Vals, in.Vals) {
		t.Errorf("typed pipe round trip: got %v want %v", out.Vals, in.Vals)
	}
}
