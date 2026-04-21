package fek_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/blob"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/headers"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/magic"
)

func makeSampleFEK(t *testing.T) fek.FEK_KEY_MATERIAL {
	t.Helper()

	modulus := make([]byte, 256)
	for i := 0; i < len(modulus); i++ {
		modulus[i] = byte(i)
	}
	aesKDFKey := make([]byte, 32)
	for i := 0; i < len(aesKDFKey); i++ {
		aesKDFKey[i] = byte(0xA0 + i)
	}

	return fek.FEK_KEY_MATERIAL{
		Version: magic.FEK_KEY_MATERIAL_VERSION{Version: magic.FEK_KEY_VERSION_1},
		Header: headers.FEK_KEY_MATERIAL_HEADER{
			BitLength:   2048,
			CbPublicExp: 3,
			CbModulus:   256,
			CbAESKDFKey: 32,
		},
		Content: blob.FEK_KEY_MATERIAL_BLOB{
			PublicExponent: []byte{0x01, 0x00, 0x01},
			Modulus:        modulus,
			AESKDFKey:      aesKDFKey,
		},
	}
}

func TestFEK_KEY_MATERIAL_MarshalUnmarshalRoundtrip(t *testing.T) {
	source := makeSampleFEK(t)

	raw, err := source.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Expected size: 1 byte version + 16 bytes header + exp + modulus + AES KDF key
	expectedLen := 1 + 16 + len(source.Content.PublicExponent) + len(source.Content.Modulus) + len(source.Content.AESKDFKey)
	if len(raw) != expectedLen {
		t.Errorf("Marshal produced wrong size, got %d, want %d", len(raw), expectedLen)
	}

	parsed := fek.FEK_KEY_MATERIAL{}
	n, err := parsed.Unmarshal(raw)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != len(raw) {
		t.Errorf("Unmarshal consumed wrong size, got %d, want %d", n, len(raw))
	}
	if !parsed.Equal(&source) {
		t.Errorf("Roundtripped FEK_KEY_MATERIAL does not match source")
	}

	if !bytes.Equal(parsed.Content.PublicExponent, source.Content.PublicExponent) {
		t.Errorf("PublicExponent mismatch")
	}
	if !bytes.Equal(parsed.Content.Modulus, source.Content.Modulus) {
		t.Errorf("Modulus mismatch")
	}
	if !bytes.Equal(parsed.Content.AESKDFKey, source.Content.AESKDFKey) {
		t.Errorf("AESKDFKey mismatch")
	}
}

func TestFEK_KEY_MATERIAL_Unmarshal_ShortInput(t *testing.T) {
	k := fek.FEK_KEY_MATERIAL{}
	_, err := k.Unmarshal([]byte{0x01, 0x00, 0x00, 0x00})
	if err == nil {
		t.Fatal("Expected error on short buffer, got nil")
	}
}

func TestFEK_KEY_MATERIAL_Unmarshal_InvalidVersion(t *testing.T) {
	source := makeSampleFEK(t)
	raw, err := source.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Corrupt the version
	raw[0] = 0xFF

	k := fek.FEK_KEY_MATERIAL{}
	_, err = k.Unmarshal(raw)
	if err == nil {
		t.Fatal("Expected error on invalid version, got nil")
	}
}

func TestFEK_KEY_MATERIAL_Marshal_SetsVersion(t *testing.T) {
	source := fek.FEK_KEY_MATERIAL{
		// Deliberately do not set Version.Version — Marshal should default it to FEK_KEY_VERSION_1.
		Header: headers.FEK_KEY_MATERIAL_HEADER{
			BitLength:   2048,
			CbPublicExp: 3,
			CbModulus:   0,
			CbAESKDFKey: 0,
		},
		Content: blob.FEK_KEY_MATERIAL_BLOB{
			PublicExponent: []byte{0x01, 0x00, 0x01},
		},
	}

	raw, err := source.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if raw[0] != magic.FEK_KEY_VERSION_1 {
		t.Errorf("Marshal did not write FEK_KEY_VERSION_1 at byte 0, got 0x%02x", raw[0])
	}
}

func TestFEK_KEY_MATERIAL_Fingerprint_Differs(t *testing.T) {
	a := makeSampleFEK(t)
	b := makeSampleFEK(t)
	// Mutate b's AES KDF key — fingerprints should diverge.
	b.Content.AESKDFKey = bytes.Repeat([]byte{0xFF}, 32)

	if a.Fingerprint() == b.Fingerprint() {
		t.Errorf("Fingerprints should differ when AESKDFKey differs")
	}
}
