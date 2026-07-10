package header

import (
	"bytes"
	"testing"
)

func TestHeaderMarshalUnmarshal(t *testing.T) {
	orig := Header{
		Identifier: 0x1a2b,
		Flags:      FlagQR | FlagC | FlagT,
		QDCount:    1,
		ANCount:    2,
		NSCount:    3,
		ARCount:    4,
	}

	data, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(data) != HeaderSize {
		t.Fatalf("Marshal returned %d bytes, want %d", len(data), HeaderSize)
	}

	var decoded Header
	n, err := decoded.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != HeaderSize {
		t.Errorf("Unmarshal read %d bytes, want %d", n, HeaderSize)
	}
	if orig.Identifier != decoded.Identifier {
		t.Errorf("Identifier: got 0x%x, want 0x%x", decoded.Identifier, orig.Identifier)
	}
	if orig.Flags != decoded.Flags {
		t.Errorf("Flags: got %#x, want %#x", decoded.Flags, orig.Flags)
	}
	if orig.QDCount != decoded.QDCount {
		t.Errorf("QDCount: got %d, want %d", decoded.QDCount, orig.QDCount)
	}
	if orig.ANCount != decoded.ANCount {
		t.Errorf("ANCount: got %d, want %d", decoded.ANCount, orig.ANCount)
	}
	if orig.NSCount != decoded.NSCount {
		t.Errorf("NSCount: got %d, want %d", decoded.NSCount, orig.NSCount)
	}
	if orig.ARCount != decoded.ARCount {
		t.Errorf("ARCount: got %d, want %d", decoded.ARCount, orig.ARCount)
	}
}

func TestHeaderMarshalOutput(t *testing.T) {
	h := Header{
		Identifier: 0x44aa,
		Flags:      FlagQR | FlagTC,
		QDCount:    0x0042,
		ANCount:    0x0123,
		NSCount:    0,
		ARCount:    0xabcd,
	}
	expected := []byte{
		0x44, 0xaa, // Identifier
		0x82, 0x00, // Flags (QR | TC => 0x8000 | 0x0200 = 0x8200)
		0x00, 0x42, // QDCount
		0x01, 0x23, // ANCount
		0x00, 0x00, // NSCount
		0xab, 0xcd, // ARCount
	}
	result, err := h.Marshal()
	if err != nil {
		t.Fatal("Header.Marshal failed:", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal = % x, want % x", result, expected)
	}
}

func TestHeaderUnmarshalInvalidLength(t *testing.T) {
	var h Header
	short := []byte{0, 1, 2}
	n, err := h.Unmarshal(short)
	if err == nil {
		t.Errorf("Unmarshal should fail for too short input, got nil error")
	}
	if n != 0 {
		t.Errorf("Unmarshal read %d bytes for invalid input, want 0", n)
	}
}

func TestHeaderDescribe(t *testing.T) {
	h := &Header{
		Identifier: 555,
		Flags:      FlagQR | FlagTC,
		QDCount:    2,
		ANCount:    3,
		NSCount:    4,
		ARCount:    5,
	}
	// Just make sure calling Describe doesn't panic and prints something.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Describe panicked: %v", r)
		}
	}()
	h.Describe(1)
}
