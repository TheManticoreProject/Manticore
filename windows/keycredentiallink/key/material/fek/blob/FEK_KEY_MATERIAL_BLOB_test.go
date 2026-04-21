package blob_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/blob"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/material/fek/headers"
)

func TestFEK_KEY_MATERIAL_BLOB_MarshalUnmarshal(t *testing.T) {
	header := headers.FEK_KEY_MATERIAL_HEADER{
		BitLength:   2048,
		CbPublicExp: 3,
		CbModulus:   256,
		CbAESKDFKey: 32,
	}
	modulus := make([]byte, header.CbModulus)
	for i := 0; i < len(modulus); i++ {
		modulus[i] = byte(i)
	}
	aesKDFKey := make([]byte, header.CbAESKDFKey)
	for i := 0; i < len(aesKDFKey); i++ {
		aesKDFKey[i] = byte(0xA0 + i)
	}

	source := &blob.FEK_KEY_MATERIAL_BLOB{
		PublicExponent: []byte{0x01, 0x00, 0x01},
		Modulus:        modulus,
		AESKDFKey:      aesKDFKey,
	}

	data, err := source.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	expectedLen := len(source.PublicExponent) + len(source.Modulus) + len(source.AESKDFKey)
	if len(data) != expectedLen {
		t.Errorf("Marshal produced wrong size, got %d, want %d", len(data), expectedLen)
	}

	parsed := blob.FEK_KEY_MATERIAL_BLOB{}
	n, err := parsed.Unmarshal(header, data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Unmarshal consumed wrong size, got %d, want %d", n, len(data))
	}
	if !bytes.Equal(parsed.PublicExponent, source.PublicExponent) {
		t.Errorf("PublicExponent mismatch: got %v, want %v", parsed.PublicExponent, source.PublicExponent)
	}
	if !bytes.Equal(parsed.Modulus, source.Modulus) {
		t.Errorf("Modulus mismatch")
	}
	if !bytes.Equal(parsed.AESKDFKey, source.AESKDFKey) {
		t.Errorf("AESKDFKey mismatch")
	}
	if !parsed.Equal(source) {
		t.Errorf("Equal() returned false for roundtripped blob")
	}
}

func TestFEK_KEY_MATERIAL_BLOB_Unmarshal_ShortInput_Modulus(t *testing.T) {
	header := headers.FEK_KEY_MATERIAL_HEADER{
		BitLength:   2048,
		CbPublicExp: 3,
		CbModulus:   256,
		CbAESKDFKey: 32,
	}
	b := blob.FEK_KEY_MATERIAL_BLOB{}
	_, err := b.Unmarshal(header, []byte{0x01, 0x00, 0x01})
	if err == nil {
		t.Fatal("Expected error on short buffer for modulus, got nil")
	}
}

func TestFEK_KEY_MATERIAL_BLOB_Unmarshal_ShortInput_AESKDFKey(t *testing.T) {
	header := headers.FEK_KEY_MATERIAL_HEADER{
		BitLength:   2048,
		CbPublicExp: 3,
		CbModulus:   4,
		CbAESKDFKey: 32,
	}
	// Enough for exp+modulus but not AES KDF key
	buf := make([]byte, 3+4)
	b := blob.FEK_KEY_MATERIAL_BLOB{}
	_, err := b.Unmarshal(header, buf)
	if err == nil {
		t.Fatal("Expected error on short buffer for AES-256 KDF key, got nil")
	}
}
