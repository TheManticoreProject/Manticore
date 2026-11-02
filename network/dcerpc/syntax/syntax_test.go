package syntax

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestNDRTransferSyntax_GoldenBytes(t *testing.T) {
	// 8a885d04-1ceb-11c9-9fe8-08002b104860, version 2.0.
	// UUID wire layout: Data1/2/3 little-endian, Data4 big-endian; then major/minor LE.
	want := []byte{
		0x04, 0x5d, 0x88, 0x8a, // A (LE)
		0xeb, 0x1c, // B (LE)
		0xc9, 0x11, // C (LE)
		0x9f, 0xe8, // D (BE)
		0x08, 0x00, 0x2b, 0x10, 0x48, 0x60, // E (BE)
		0x02, 0x00, // major = 2
		0x00, 0x00, // minor = 0
	}
	got, err := NDRTransferSyntax().Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("NDR syntax bytes:\n got %x\nwant %x", got, want)
	}
	if len(got) != Size {
		t.Errorf("len = %d, want %d", len(got), Size)
	}
}

func TestSyntaxID_RoundTrip(t *testing.T) {
	for _, s := range []SyntaxID{NDRTransferSyntax(), NDR64TransferSyntax()} {
		b, err := s.Marshal()
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		var got SyntaxID
		n, err := got.Unmarshal(b)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if n != Size {
			t.Errorf("Unmarshal consumed %d bytes, want %d", n, Size)
		}
		if !got.Equal(s) {
			t.Errorf("round trip: got %s, want %s", got, s)
		}
	}
}

func TestSyntaxID_UnmarshalTruncated(t *testing.T) {
	var s SyntaxID
	if _, err := s.Unmarshal(make([]byte, Size-1)); err == nil {
		t.Fatal("Unmarshal of truncated buffer: error = nil, want non-nil")
	}
}

func TestSyntaxID_NDRvsNDR64NotEqual(t *testing.T) {
	if NDRTransferSyntax().Equal(NDR64TransferSyntax()) {
		t.Error("NDR and NDR64 syntaxes compare equal")
	}
}

// TestUnmarshalOrder_BigEndian checks that a big-endian syntax identifier decodes to
// the same SyntaxID as its little-endian form: the three integer UUID fields and the
// two version halves are byte-swapped on the wire, the clock_seq and node octets are
// not.
func TestUnmarshalOrder_BigEndian(t *testing.T) {
	ndr := NDRTransferSyntax()
	le, err := ndr.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Re-encode big-endian: swap the integer fields, leave bytes 8..16 in place.
	be := append([]byte(nil), le...)
	swap := func(b []byte) {
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}
	swap(be[0:4])
	swap(be[4:6])
	swap(be[6:8])
	swap(be[16:18])
	swap(be[18:20])

	var s SyntaxID
	n, err := s.UnmarshalOrder(be, binary.BigEndian)
	if err != nil {
		t.Fatalf("UnmarshalOrder: %v", err)
	}
	if n != Size {
		t.Errorf("consumed %d bytes, want %d", n, Size)
	}
	if !s.Equal(ndr) {
		t.Errorf("decoded %s, want %s", s, ndr)
	}

	// The default Unmarshal stays little-endian: it reads the big-endian bytes the
	// other way and must NOT match NDR.
	var leParsed SyntaxID
	if _, err := leParsed.Unmarshal(be); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if leParsed.Equal(ndr) {
		t.Error("little-endian Unmarshal of big-endian bytes matched NDR; byte order not honored")
	}
}
