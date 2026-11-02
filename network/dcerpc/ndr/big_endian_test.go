package ndr

import (
	"encoding/binary"
	"testing"
)

// TestDecoder_ByteOrder checks the decoder reads multi-octet integers in the byte
// order it was given, so the same octets decode to different values under the two
// orders. This is the whole surface of big-endian NDR support: every typed read
// funnels through ReadUint16/32/64.
func TestDecoder_ByteOrder(t *testing.T) {
	octets := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}

	be := NewDecoderForSyntaxWithOrder(octets, NDR20, binary.BigEndian)
	if v, err := be.ReadUint16(); err != nil || v != 0x1122 {
		t.Errorf("big-endian ReadUint16 = %#x, %v; want 0x1122", v, err)
	}
	be = NewDecoderForSyntaxWithOrder(octets, NDR20, binary.BigEndian)
	if v, err := be.ReadUint32(); err != nil || v != 0x11223344 {
		t.Errorf("big-endian ReadUint32 = %#x, %v; want 0x11223344", v, err)
	}
	be = NewDecoderForSyntaxWithOrder(octets, NDR20, binary.BigEndian)
	if v, err := be.ReadUint64(); err != nil || v != 0x1122334455667788 {
		t.Errorf("big-endian ReadUint64 = %#x, %v; want 0x1122334455667788", v, err)
	}

	// The default decoder (and an explicit little-endian one) read the same octets
	// least-significant byte first — unchanged from before.
	le := NewDecoder(octets)
	if v, err := le.ReadUint32(); err != nil || v != 0x44332211 {
		t.Errorf("little-endian ReadUint32 = %#x, %v; want 0x44332211", v, err)
	}
	if NewDecoderForSyntaxWithOrder(octets, NDR20, nil).byteOrder() != binary.LittleEndian {
		t.Error("a nil order does not default to little-endian")
	}
}

// TestUnmarshalAsOrder_BigEndian checks the order reaches the struct decode path:
// a dword whose octets are 0x11 0x22 0x33 0x44 decodes to 0x11223344 big-endian and
// to 0x44332211 little-endian.
func TestUnmarshalAsOrder_BigEndian(t *testing.T) {
	type one struct {
		V uint32 `ndr:"dword"`
	}
	octets := []byte{0x11, 0x22, 0x33, 0x44}

	var be one
	if err := UnmarshalAsOrder(octets, &be, NDR20, binary.BigEndian); err != nil {
		t.Fatalf("UnmarshalAsOrder big-endian: %v", err)
	}
	if be.V != 0x11223344 {
		t.Errorf("big-endian V = %#x, want 0x11223344", be.V)
	}

	var le one
	if err := UnmarshalAsOrder(octets, &le, NDR20, binary.LittleEndian); err != nil {
		t.Fatalf("UnmarshalAsOrder little-endian: %v", err)
	}
	if le.V != 0x44332211 {
		t.Errorf("little-endian V = %#x, want 0x44332211", le.V)
	}

	// The long-standing little-endian entry points are unchanged.
	var def one
	if err := Unmarshal(octets, &def); err != nil || def.V != 0x44332211 {
		t.Errorf("Unmarshal default V = %#x, %v; want 0x44332211", def.V, err)
	}
}

// TestDecoder_WStringByteOrder checks a UTF-16 string's code units are read in the
// decoder's byte order: wide strings are the most common MS-RPC field, and their
// code units are 2-octet integers subject to the data representation.
func TestDecoder_WStringByteOrder(t *testing.T) {
	// A conformant+varying WSTR "Hi": max_count, offset, actual_count (3, incl. the
	// NUL), then the code units H(0x0048) i(0x0069) NUL(0x0000).
	be := []byte{
		0, 0, 0, 3, // max_count
		0, 0, 0, 0, // offset
		0, 0, 0, 3, // actual_count
		0x00, 0x48, 0x00, 0x69, 0x00, 0x00, // big-endian code units
	}
	if s, err := NewDecoderForSyntaxWithOrder(be, NDR20, binary.BigEndian).readWString(); err != nil || s != "Hi" {
		t.Errorf("big-endian readWString = %q, %v; want %q", s, err, "Hi")
	}

	le := []byte{
		3, 0, 0, 0,
		0, 0, 0, 0,
		3, 0, 0, 0,
		0x48, 0x00, 0x69, 0x00, 0x00, 0x00, // little-endian code units
	}
	if s, err := NewDecoder(le).readWString(); err != nil || s != "Hi" {
		t.Errorf("little-endian readWString = %q, %v; want %q", s, err, "Hi")
	}
}
