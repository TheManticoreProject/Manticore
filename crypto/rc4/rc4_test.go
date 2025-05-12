package rc4_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/rc4"
)

func TestRC4Cipher(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		plaintext string
		expected  string
	}{
		{
			name:      "RFC 6229 Test Vector (40-bit key)",
			key:       "0102030405",
			plaintext: "00000000000000000000000000000000",
			expected:  "b2396305f03dc027ccc3524a0a1118a8",
		},
		{
			name:      "RFC 6229 Test Vector (56-bit key)",
			key:       "01020304050607",
			plaintext: "00000000000000000000000000000000",
			expected:  "293f02d47f37c9b633f2af5285feb46b",
		},
		{
			name:      "RFC 6229 Test Vector (128-bit key)",
			key:       "0102030405060708090a0b0c0d0e0f10",
			plaintext: "00000000000000000000000000000000",
			expected:  "9ac7cc9a609d1ef7b2932899cde41b97",
		},
		{
			name:      "Simple Text Encryption",
			key:       "506f64616c6972697573",
			plaintext: "Hello, World!",
			expected:  "cc18ccc14fecc0bd66acd1b5ee",
		},
		{
			name:      "One null byte Key",
			key:       "00",
			plaintext: "Hello, World!",
			expected:  "967de52dcc1b7d6de574720376",
		},
		{
			name:      "Two null bytes Key",
			key:       "0000",
			plaintext: "Hello, World!",
			expected:  "967de52dcc1b7d6de574720376",
		},
		{
			name:      "Three null bytes Key",
			key:       "000000",
			plaintext: "Hello, World!",
			expected:  "967de52dcc1b7d6de574720376",
		},
		{
			name:      "Four null bytes Key",
			key:       "00000000",
			plaintext: "Hello, World!",
			expected:  "967de52dcc1b7d6de574720376",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _ := hex.DecodeString(tt.key)
			if _, err := hex.DecodeString(tt.key); err != nil {
				// If not valid hex, use as raw bytes
				key = []byte(tt.key)
			}

			plaintext, _ := hex.DecodeString(tt.plaintext)
			if _, err := hex.DecodeString(tt.plaintext); err != nil {
				// If not valid hex, use as raw bytes
				plaintext = []byte(tt.plaintext)
			}

			expected, _ := hex.DecodeString(tt.expected)

			cipher, err := rc4.NewRC4WithKey(key)
			if err != nil {
				t.Fatalf("Failed to create RC4 cipher: %v", err)
			}

			result := make([]byte, len(plaintext))
			cipher.XORKeyStream(result, plaintext)

			if !bytes.Equal(result, expected) {
				t.Errorf("Expected: %x, got: %x", expected, result)
			}

			// Test decryption (RC4 is symmetric)
			cipher, _ = rc4.NewRC4WithKey(key)
			decrypted := make([]byte, len(result))
			cipher.XORKeyStream(decrypted, result)

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("Decryption failed. Expected: %x, got: %x", plaintext, decrypted)
			}
		})
	}
}

func TestInvolution(t *testing.T) {
	testStrings := []string{
		"Hello, World!",
		"The quick brown fox jumps over the lazy dog",
		"",
		"1234567890",
		"RC4 is a stream cipher designed by Ron Rivest in 1987",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
	}

	keys := [][]byte{
		[]byte("secret"),
		[]byte("0102030405"),
		[]byte("ThisIsALongerKeyForTesting"),
		[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		[]byte("Podalirius"),
	}

	for _, key := range keys {
		for _, plaintext := range testStrings {
			t.Run(fmt.Sprintf("Key:%x-Text:%s", key, plaintext[:min(len(plaintext), 10)]), func(t *testing.T) {
				// Create a new RC4 cipher with the key
				cipher1, err := rc4.NewRC4WithKey(key)
				if err != nil {
					t.Fatalf("Failed to create RC4 cipher: %v", err)
				}

				// Encrypt the plaintext
				plaintextBytes := []byte(plaintext)
				encrypted := make([]byte, len(plaintextBytes))
				cipher1.XORKeyStream(encrypted, plaintextBytes)

				// Create a new RC4 cipher with the same key for decryption
				cipher2, err := rc4.NewRC4WithKey(key)
				if err != nil {
					t.Fatalf("Failed to create RC4 cipher for decryption: %v", err)
				}

				// Decrypt the ciphertext
				decrypted := make([]byte, len(encrypted))
				cipher2.XORKeyStream(decrypted, encrypted)

				// Verify that the decrypted text matches the original plaintext
				if !bytes.Equal(plaintextBytes, decrypted) {
					t.Errorf("Involution property failed. Original: %q, Decrypted: %q", plaintext, string(decrypted))
				}
			})
		}
	}
}

// min returns the smaller of x or y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func TestRC4Decrypt(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		ciphertext string
		expected   string
	}{
		{
			name:       "RFC 6229 Test Vector (40-bit key)",
			key:        "0102030405",
			expected:   "00000000000000000000000000000000",
			ciphertext: "b2396305f03dc027ccc3524a0a1118a8",
		},
		{
			name:       "RFC 6229 Test Vector (56-bit key)",
			key:        "01020304050607",
			expected:   "00000000000000000000000000000000",
			ciphertext: "293f02d47f37c9b633f2af5285feb46b",
		},
		{
			name:       "RFC 6229 Test Vector (128-bit key)",
			key:        "0102030405060708090a0b0c0d0e0f10",
			expected:   "00000000000000000000000000000000",
			ciphertext: "9ac7cc9a609d1ef7b2932899cde41b97",
		},
		{
			name:       "Simple Text Encryption",
			key:        "506f64616c6972697573",
			expected:   "48656c6c6f2c20576f726c6421",
			ciphertext: "cc18ccc14fecc0bd66acd1b5ee",
		},
		{
			name:       "One null byte Key",
			key:        "00",
			expected:   "48656c6c6f2c20576f726c6421",
			ciphertext: "967de52dcc1b7d6de574720376",
		},
		{
			name:       "Two null bytes Key",
			key:        "0000",
			expected:   "48656c6c6f2c20576f726c6421",
			ciphertext: "967de52dcc1b7d6de574720376",
		},
		{
			name:       "Three null bytes Key",
			key:        "000000",
			expected:   "48656c6c6f2c20576f726c6421",
			ciphertext: "967de52dcc1b7d6de574720376",
		},
		{
			name:       "Four null bytes Key",
			key:        "00000000",
			expected:   "48656c6c6f2c20576f726c6421",
			ciphertext: "967de52dcc1b7d6de574720376",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _ := hex.DecodeString(tt.key)
			if _, err := hex.DecodeString(tt.key); err != nil {
				// If not valid hex, use as raw bytes
				key = []byte(tt.key)
			}

			ciphertext, _ := hex.DecodeString(tt.ciphertext)
			if _, err := hex.DecodeString(tt.ciphertext); err != nil {
				// If not valid hex, use as raw bytes
				ciphertext = []byte(tt.ciphertext)
			}

			expected, _ := hex.DecodeString(tt.expected)

			cipher, err := rc4.NewRC4WithKey(key)
			if err != nil {
				t.Fatalf("Failed to create RC4 cipher: %v", err)
			}

			result := make([]byte, len(ciphertext))
			cipher.XORKeyStream(result, ciphertext)

			if !bytes.Equal(result, expected) {
				t.Errorf("Expected: %x, got: %x", expected, result)
			}

			// Test encryption (RC4 is symmetric)
			cipher, _ = rc4.NewRC4WithKey(key)
			encrypted := make([]byte, len(result))
			cipher.XORKeyStream(encrypted, result)

			if !bytes.Equal(encrypted, ciphertext) {
				t.Errorf("Re-encryption failed. Expected: %x, got: %x", ciphertext, encrypted)
			}
		})
	}
}

func TestRC4InvalidKey(t *testing.T) {
	// Test with empty key
	_, err := rc4.NewRC4()
	if err == nil {
		t.Error("Expected error with empty key, got nil")
	}

	// Test with key that's too long
	longKey := make([]byte, 257)
	_, err = rc4.NewRC4WithKey(longKey)
	if err == nil {
		t.Error("Expected error with key > 256 bytes, got nil")
	}
}

func TestRC4Reset(t *testing.T) {
	key := []byte("TestKey123")
	plaintext := []byte("Hello, World!")

	cipher, _ := rc4.NewRC4WithKey(key)
	result1 := make([]byte, len(plaintext))
	cipher.XORKeyStream(result1, plaintext)

	// Reset the cipher
	cipher.Reset()

	result2 := make([]byte, len(plaintext))
	cipher.XORKeyStream(result2, plaintext)

	// After reset, the cipher should not produce same output
	// because the state has been zeroed
	if bytes.Equal(result1, result2) {
		t.Error("Reset did not change cipher state as expected")
		t.Errorf("Result1: %x", result1)
		t.Errorf("Result2: %x", result2)
	}
}
