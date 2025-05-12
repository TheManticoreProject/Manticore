package lm

import (
	"crypto/des"
	"encoding/hex"
	"strings"
)

// LMHash computes the LAN Manager hash of a password string
// The LM hash is computed as follows:
// 1. Convert password to uppercase and pad/truncate to 14 bytes
// 2. Split into two 7-byte halves
// 3. Create two DES keys from each 7-byte half (convert 7 bytes to 8 bytes adding parity bits)
// 4. DES encrypt the constant "KGS!@#$%" with each key
// 5. Concatenate the two DES outputs
func LMHash(password string) []byte {
	// Convert to uppercase and pad/truncate to 14 bytes
	password = strings.ToUpper(password)
	if len(password) > 14 {
		password = password[:14]
	}
	if len(password) < 14 {
		password = password + strings.Repeat("\x00", 14-len(password))
	}

	// Split into two 7-byte halves
	firstHalf := []byte(password[:7])
	secondHalf := []byte(password[7:])

	// Convert 7-byte halves to 8-byte DES keys (with parity bits)
	firstKey := make([]byte, 8)
	secondKey := make([]byte, 8)

	// Add parity bits
	firstKey[0] = firstHalf[0]
	firstKey[1] = (firstHalf[0] << 7) | (firstHalf[1] >> 1)
	firstKey[2] = (firstHalf[1] << 6) | (firstHalf[2] >> 2)
	firstKey[3] = (firstHalf[2] << 5) | (firstHalf[3] >> 3)
	firstKey[4] = (firstHalf[3] << 4) | (firstHalf[4] >> 4)
	firstKey[5] = (firstHalf[4] << 3) | (firstHalf[5] >> 5)
	firstKey[6] = (firstHalf[5] << 2) | (firstHalf[6] >> 6)
	firstKey[7] = (firstHalf[6] << 1)

	secondKey[0] = secondHalf[0]
	secondKey[1] = (secondHalf[0] << 7) | (secondHalf[1] >> 1)
	secondKey[2] = (secondHalf[1] << 6) | (secondHalf[2] >> 2)
	secondKey[3] = (secondHalf[2] << 5) | (secondHalf[3] >> 3)
	secondKey[4] = (secondHalf[3] << 4) | (secondHalf[4] >> 4)
	secondKey[5] = (secondHalf[4] << 3) | (secondHalf[5] >> 5)
	secondKey[6] = (secondHalf[5] << 2) | (secondHalf[6] >> 6)
	secondKey[7] = (secondHalf[6] << 1)

	// The constant "KGS!@#$%" to encrypt
	magic := []byte("KGS!@#$%")

	// Create DES ciphers
	cipher1, _ := des.NewCipher(firstKey)
	cipher2, _ := des.NewCipher(secondKey)

	// Encrypt the magic constant with both keys
	result1 := make([]byte, 8)
	result2 := make([]byte, 8)
	cipher1.Encrypt(result1, magic)
	cipher2.Encrypt(result2, magic)

	// Concatenate results
	finalHash := append(result1, result2...)

	// Convert to lowercase hex string
	return finalHash
}

// LMHashToHex converts the LM hash to a lowercase hex string
//
// # The LM hash is converted to a hex string by encoding the hash as a lowercase hex string
//
// Returns the lowercase hex string representation of the LM hash
func LMHashToHex(password string) string {
	return strings.ToLower(hex.EncodeToString(LMHash(password)))
}
