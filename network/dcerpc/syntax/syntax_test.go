package syntax

import (
	"bytes"
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
