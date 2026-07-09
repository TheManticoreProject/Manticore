package pac

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// sigBuf builds a zero-signature PAC_SIGNATURE_DATA buffer for the given
// SignatureType and checksum size.
func sigBuf(sigType uint32, size int) []byte {
	b := make([]byte, 4+size)
	binary.LittleEndian.PutUint32(b, sigType)
	return b
}

func buildSignedPAC(t *testing.T, serverKey, kdcKey []byte) []byte {
	t.Helper()
	p := &PAC{
		Buffers: []Buffer{
			{Type: BufferLogonInfo, Data: bytes.Repeat([]byte{0xAA}, 40)},
			{Type: BufferClientInfo, Data: []byte("client-info-bytes")},
			{Type: BufferServerChecksum, Data: sigBuf(sigHMACSHA1256, 12)},
			{Type: BufferKDCChecksum, Data: sigBuf(sigHMACSHA1256, 12)},
		},
	}
	signed, err := p.Sign(serverKey, kdcKey)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return signed
}

func TestPACSignVerifyRoundtrip(t *testing.T) {
	serverKey := bytes.Repeat([]byte{0x11}, 32)
	kdcKey := bytes.Repeat([]byte{0x22}, 32)

	signed := buildSignedPAC(t, serverKey, kdcKey)

	p, err := Parse(signed)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if p.Version != 0 || len(p.Buffers) != 4 {
		t.Fatalf("parsed PAC wrong: version=%d buffers=%d", p.Version, len(p.Buffers))
	}
	if _, ok := p.Buffer(BufferLogonInfo); !ok {
		t.Error("logon info buffer missing")
	}

	if err := p.VerifyServerSignature(serverKey); err != nil {
		t.Errorf("server signature should verify: %v", err)
	}
	if err := p.VerifyKDCSignature(kdcKey); err != nil {
		t.Errorf("KDC signature should verify: %v", err)
	}
}

func TestPACWrongKeyFails(t *testing.T) {
	serverKey := bytes.Repeat([]byte{0x11}, 32)
	kdcKey := bytes.Repeat([]byte{0x22}, 32)
	signed := buildSignedPAC(t, serverKey, kdcKey)

	p, _ := Parse(signed)
	if err := p.VerifyServerSignature(bytes.Repeat([]byte{0x99}, 32)); err == nil {
		t.Error("server signature verified under the wrong key")
	}
	if err := p.VerifyKDCSignature(bytes.Repeat([]byte{0x99}, 32)); err == nil {
		t.Error("KDC signature verified under the wrong key")
	}
}

func TestPACTamperFails(t *testing.T) {
	serverKey := bytes.Repeat([]byte{0x11}, 32)
	kdcKey := bytes.Repeat([]byte{0x22}, 32)
	signed := buildSignedPAC(t, serverKey, kdcKey)

	// Flip a byte inside the logon-info buffer (which lies after the header/table).
	tampered := append([]byte(nil), signed...)
	tampered[8+4*16+0] ^= 0x01 // first byte of the first 8-aligned buffer region

	p, err := Parse(tampered)
	if err != nil {
		t.Fatalf("Parse tampered: %v", err)
	}
	if err := p.VerifyServerSignature(serverKey); err == nil {
		t.Error("server signature verified over tampered PAC data")
	}
}

func TestPACParseRejectsGarbage(t *testing.T) {
	if _, err := Parse([]byte{1, 2, 3}); err == nil {
		t.Error("expected error for short PAC")
	}
	// version != 0
	bad := make([]byte, 8)
	binary.LittleEndian.PutUint32(bad[0:], 0)
	binary.LittleEndian.PutUint32(bad[4:], 1)
	if _, err := Parse(bad); err == nil {
		t.Error("expected error for non-zero version")
	}
	// buffer offset out of range
	oob := make([]byte, 8+16)
	binary.LittleEndian.PutUint32(oob[0:], 1)  // cBuffers
	binary.LittleEndian.PutUint32(oob[4:], 0)  // version
	binary.LittleEndian.PutUint32(oob[8:], 1)  // ulType
	binary.LittleEndian.PutUint32(oob[12:], 8) // cbBufferSize
	binary.LittleEndian.PutUint64(oob[16:], 9999)
	if _, err := Parse(oob); err == nil {
		t.Error("expected error for out-of-range buffer offset")
	}
}

func TestPACMarshalParseBufferRoundtrip(t *testing.T) {
	p := &PAC{Buffers: []Buffer{
		{Type: BufferLogonInfo, Data: []byte("logon")},
		{Type: BufferUPNDNSInfo, Data: []byte("upn-dns-info-longer")},
	}}
	m, err := p.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Buffers) != 2 {
		t.Fatalf("buffers: %d", len(got.Buffers))
	}
	b0, _ := got.Buffer(BufferLogonInfo)
	b1, _ := got.Buffer(BufferUPNDNSInfo)
	if string(b0.Data) != "logon" || string(b1.Data) != "upn-dns-info-longer" {
		t.Errorf("buffer data round-trip mismatch: %q %q", b0.Data, b1.Data)
	}
	// Offsets must be 8-aligned.
	for _, b := range got.Buffers {
		if b.Offset%8 != 0 {
			t.Errorf("buffer type 0x%x offset %d not 8-aligned", b.Type, b.Offset)
		}
	}
}
