package kerbcrypto

import (
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/crypto/aescts"
	"github.com/TheManticoreProject/Manticore/crypto/nfold"
	"golang.org/x/crypto/pbkdf2"
)

// aesDefaultIterCount is the default PBKDF2 iteration count for AES string-to-key
// as specified in RFC 3962 Section 4.
const aesDefaultIterCount = 4096

// deriveKey implements the RFC 3961 DR (Derived Random) function and extracts
// a key of key_len bytes by repeatedly AES-encrypting the n-folded constant.
//
// DK(base_key, constant) = first key_len bytes of:
//
//	AES-ECB(base_key, n-fold(constant, 128))
//	AES-ECB(base_key, previous_block)
//	...
func deriveKey(base_key []byte, constant []byte, key_len int) []byte {
	// N-fold the constant to 128 bits (16 bytes)
	nfolded := nfold.NFold(constant, 128)

	block_cipher, err := aes.NewCipher(base_key)
	if err != nil {
		// Key length error; caller should have validated
		return nil
	}

	result := make([]byte, 0, key_len)
	// Use n as the running AES input, starting from the n-folded value
	n := make([]byte, 16)
	copy(n, nfolded)
	for len(result) < key_len {
		block_cipher.Encrypt(n, n) // AES-ECB: encrypt in place
		result = append(result, n...)
	}
	return result[:key_len]
}

// usageConstant builds the 5-byte usage constant used for AES key derivation.
// Format: 4-byte big-endian usage number || 1-byte purpose flag.
// Purpose flags: 0xAA = encryption, 0x55 = integrity, 0x99 = checksum.
func usageConstant(usage int, purpose byte) []byte {
	b := make([]byte, 5)
	binary.BigEndian.PutUint32(b, uint32(usage))
	b[4] = purpose
	return b
}

// aesStringToKey derives an AES key from a password and salt using PBKDF2,
// then applies the RFC 3961 random-to-key function.
// Per RFC 3962 Section 4.
func aesStringToKey(password, salt string, iter_count, key_len int) ([]byte, error) {
	// PBKDF2-HMAC-SHA1
	tkey := pbkdf2.Key([]byte(password), []byte(salt), iter_count, key_len, sha1.New)
	// Apply DK with the constant "kerberos"
	dk := deriveKey(tkey, []byte("kerberos"), key_len)
	return dk, nil
}

// aesKeyLen returns the key length in bytes for an AES Kerberos etype.
// etype 17 = AES-128 (16 bytes), etype 18 = AES-256 (32 bytes).
func aesKeyLen(etype int) int {
	if etype == 17 {
		return 16
	}
	return 32
}

// aesEncrypt encrypts plaintext using AES-CTS-HMAC-SHA1-96 per RFC 3962 + RFC 3961.
//
// Process:
//  1. Derive encryption key Ke = DK(key, usage||0xAA)
//  2. Derive integrity key Ki = DK(key, usage||0x55)
//  3. Generate 16-byte random confounder
//  4. plaintext_with_conf = confounder || plaintext
//  5. ciphertext = AES-CTS(Ke, zero_iv, plaintext_with_conf)
//  6. mac = HMAC-SHA1(Ki, plaintext_with_conf)[:12]
//  7. output = ciphertext || mac
func aesEncrypt(key []byte, etype, usage int, plaintext []byte) ([]byte, error) {
	key_len := aesKeyLen(etype)

	// Derive encryption and integrity keys
	ke := deriveKey(key, usageConstant(usage, 0xAA), key_len)
	ki := deriveKey(key, usageConstant(usage, 0x55), key_len)

	// Generate 16-byte confounder
	conf := make([]byte, 16)
	if _, err := randRead(conf); err != nil {
		return nil, err
	}

	// plaintext_with_conf = confounder || plaintext
	ptc := make([]byte, 16+len(plaintext))
	copy(ptc[:16], conf)
	copy(ptc[16:], plaintext)

	// AES-CTS encrypt with zero IV
	zero_iv := make([]byte, 16)
	ciphertext, err := aescts.Encrypt(ke, zero_iv, ptc)
	if err != nil {
		return nil, err
	}

	// HMAC-SHA1 integrity check, truncated to 12 bytes
	mac_full := hmacSHA1(ki, ptc)
	mac := mac_full[:12]

	// output = ciphertext || mac
	result := make([]byte, len(ciphertext)+12)
	copy(result[:len(ciphertext)], ciphertext)
	copy(result[len(ciphertext):], mac)
	return result, nil
}

// aesDecrypt decrypts ciphertext using AES-CTS-HMAC-SHA1-96 per RFC 3962 + RFC 3961.
//
// Input format: AES-CTS-encrypted(confounder || plaintext) || mac (12 bytes)
func aesDecrypt(key []byte, etype, usage int, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 28 { // 16 confounder + 12 mac minimum
		return nil, ErrCiphertextTooShort
	}

	key_len := aesKeyLen(etype)

	// Derive encryption and integrity keys
	ke := deriveKey(key, usageConstant(usage, 0xAA), key_len)
	ki := deriveKey(key, usageConstant(usage, 0x55), key_len)

	// Split ciphertext and MAC
	mac := ciphertext[len(ciphertext)-12:]
	enc := ciphertext[:len(ciphertext)-12]

	// AES-CTS decrypt with zero IV
	zero_iv := make([]byte, 16)
	ptc, err := aescts.Decrypt(ke, zero_iv, enc)
	if err != nil {
		return nil, err
	}

	// Verify integrity
	expected_mac := hmacSHA1(ki, ptc)[:12]
	if !hmac.Equal(mac, expected_mac) {
		return nil, ErrIntegrityCheckFailed
	}

	// Strip the 16-byte confounder
	return ptc[16:], nil
}

// hmacSHA1 computes HMAC-SHA1 of data using key.
func hmacSHA1(key, data []byte) []byte {
	mac := hmac.New(sha1.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
