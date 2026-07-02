package cfb8

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"testing"
)

// mustHex decodes a hex string or fails the test.
func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex.DecodeString(%q): %v", s, err)
	}
	return b
}

// TestCFB8AES128KnownAnswer validates the encrypter against the NIST SP 800-38A
// F.3.7 (CFB8-AES128.Encrypt) known-answer test vector, covering multiple
// segments so the shift-register feedback is exercised beyond the first octet.
func TestCFB8AES128KnownAnswer(t *testing.T) {
	key := mustHex(t, "2b7e151628aed2a6abf7158809cf4f3c")
	iv := mustHex(t, "000102030405060708090a0b0c0d0e0f")
	plaintext := mustHex(t, "6bc1bee22e409f96")
	want := mustHex(t, "3b79424c9c0dd436")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	got := make([]byte, len(plaintext))
	NewEncrypter(block, iv).XORKeyStream(got, plaintext)
	if !bytes.Equal(got, want) {
		t.Fatalf("CFB8 encrypt = %x, want %x", got, want)
	}
}

// TestCFB8RoundTrip verifies that decrypting an encrypted message with the same
// key and IV recovers the plaintext, for a length that is not a multiple of the
// block size.
func TestCFB8RoundTrip(t *testing.T) {
	key := mustHex(t, "2b7e151628aed2a6abf7158809cf4f3c")
	iv := mustHex(t, "000102030405060708090a0b0c0d0e0f")
	plaintext := []byte("cfb8 mode processes one octet per block cipher call")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	ciphertext := make([]byte, len(plaintext))
	NewEncrypter(block, iv).XORKeyStream(ciphertext, plaintext)
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext equals plaintext")
	}

	recovered := make([]byte, len(ciphertext))
	NewDecrypter(block, iv).XORKeyStream(recovered, ciphertext)
	if !bytes.Equal(recovered, plaintext) {
		t.Fatalf("round trip = %q, want %q", recovered, plaintext)
	}
}

// TestCFB8InPlace verifies that XORKeyStream works with dst aliasing src.
func TestCFB8InPlace(t *testing.T) {
	key := mustHex(t, "2b7e151628aed2a6abf7158809cf4f3c")
	iv := mustHex(t, "000102030405060708090a0b0c0d0e0f")
	buf := mustHex(t, "6bc1bee22e409f96")
	want := mustHex(t, "3b79424c9c0dd436")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	NewEncrypter(block, iv).XORKeyStream(buf, buf)
	if !bytes.Equal(buf, want) {
		t.Fatalf("in-place encrypt = %x, want %x", buf, want)
	}
}

// TestCFB8BadIVPanics verifies the constructors reject an IV whose length does
// not match the block size.
func TestCFB8BadIVPanics(t *testing.T) {
	block, err := aes.NewCipher(mustHex(t, "2b7e151628aed2a6abf7158809cf4f3c"))
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for short IV")
		}
	}()
	NewEncrypter(block, []byte{0x00})
}
