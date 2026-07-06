package cts

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// TestEncryptDecryptRoundtrip tests that Encrypt followed by Decrypt returns the original plaintext
// for various input lengths.
func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := make([]byte, 16) // AES-128
	iv := make([]byte, 16)
	rand.Read(key)
	rand.Read(iv)

	lengths := []int{16, 17, 20, 31, 32, 33, 40, 48, 64, 100}

	for _, length := range lengths {
		plaintext := make([]byte, length)
		rand.Read(plaintext)

		ciphertext, err := Encrypt(key, iv, plaintext)
		if err != nil {
			t.Errorf("Encrypt(%d bytes) error: %v", length, err)
			continue
		}

		if len(ciphertext) != len(plaintext) {
			t.Errorf("Encrypt(%d bytes) output length = %d, want %d", length, len(ciphertext), len(plaintext))
			continue
		}

		decrypted, err := Decrypt(key, iv, ciphertext)
		if err != nil {
			t.Errorf("Decrypt(%d bytes) error: %v", length, err)
			continue
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("Roundtrip(%d bytes) failed: got %x, want %x", length, decrypted, plaintext)
		}
	}
}

// TestEncryptDecryptAES256Roundtrip tests with a 256-bit key.
func TestEncryptDecryptAES256Roundtrip(t *testing.T) {
	key := make([]byte, 32) // AES-256
	iv := make([]byte, 16)
	rand.Read(key)
	rand.Read(iv)

	lengths := []int{16, 20, 32, 40}

	for _, length := range lengths {
		plaintext := make([]byte, length)
		rand.Read(plaintext)

		ciphertext, err := Encrypt(key, iv, plaintext)
		if err != nil {
			t.Errorf("AES-256 Encrypt(%d bytes) error: %v", length, err)
			continue
		}

		decrypted, err := Decrypt(key, iv, ciphertext)
		if err != nil {
			t.Errorf("AES-256 Decrypt(%d bytes) error: %v", length, err)
			continue
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("AES-256 Roundtrip(%d bytes) failed", length)
		}
	}
}

// TestEncryptTooShort verifies that Encrypt rejects plaintext shorter than 16 bytes.
func TestEncryptTooShort(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	_, err := Encrypt(key, iv, []byte("short"))
	if err == nil {
		t.Error("Encrypt with < 16 bytes should return an error")
	}
}

// TestDecryptTooShort verifies that Decrypt rejects ciphertext shorter than 16 bytes.
func TestDecryptTooShort(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	_, err := Decrypt(key, iv, []byte("short"))
	if err == nil {
		t.Error("Decrypt with < 16 bytes should return an error")
	}
}

// TestEncryptDeterministic verifies that Encrypt with the same inputs produces the same output.
func TestEncryptDeterministic(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	plaintext := make([]byte, 32)

	c1, err := Encrypt(key, iv, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Encrypt(key, iv, plaintext)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(c1, c2) {
		t.Error("Encrypt is not deterministic")
	}
}
