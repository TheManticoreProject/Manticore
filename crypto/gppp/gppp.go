package gppp

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/pkcs7"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// AES Key as per the Microsoft documentation
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-gppref/2c15cbf0-f086-4c74-8b70-1f2fa45dd4be
var GPPP_AES_KEY = []byte{
	0x4e, 0x99, 0x06, 0xe8, 0xfc, 0xb6, 0x6c, 0xc9, 0xfa, 0xf4, 0x93, 0x10, 0x62, 0x0f, 0xfe, 0xe8,
	0xf4, 0x96, 0xe8, 0x06, 0xcc, 0x05, 0x79, 0x90, 0x20, 0x9b, 0x09, 0xa4, 0x33, 0xb6, 0x6c, 0x1b,
}

// GPPPDecrypt decrypts a base64 encoded string using the fixed AES key and IV
//
// Parameters:
// - encStr: The base64 encoded string to be decrypted.
//
// Returns:
// - The decrypted string.
// - An error if the decryption or unpadding fails.
func GPPPDecryptBase64(encStr string) (string, error) {
	// Padding base64 encoded string to ensure it's properly padded
	pad := len(encStr) % 4
	if pad == 1 {
		encStr = encStr[:len(encStr)-1]
	} else if pad == 2 || pad == 3 {
		encStr += strings.Repeat("=", 4-pad)
	}

	// Decode base64 string
	ciphertext, err := base64.StdEncoding.DecodeString(encStr)
	if err != nil {
		return "", fmt.Errorf("base64 decoding failed: %v", err)
	}

	return GPPPDecryptBytes(ciphertext)
}

// GPPPDecryptBytes decrypts a byte slice using the fixed AES key and IV, and returns the decrypted string.
//
// Parameters:
// - ciphertext: The byte slice to be decrypted.
//
// Returns:
// - The decrypted string.
// - An error if the decryption or unpadding fails.
//
// The function performs the following steps:
// 1. Initializes a fixed null IV (Initialization Vector).
// 2. Creates an AES cipher block using the fixed AES key.
// 3. Ensures the ciphertext length is a multiple of the AES block size.
// 4. Creates a CBC decrypter and decrypts the ciphertext.
// 5. Removes PKCS#7 padding from the decrypted plaintext.
// 6. Converts the plaintext from UTF-16LE to a string and returns it.
func GPPPDecryptBytes(ciphertext []byte) (string, error) {
	// Fixed null IV (Initialization Vector)
	iv := make([]byte, aes.BlockSize)

	// Create AES cipher block
	block, err := aes.NewCipher(GPPP_AES_KEY)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// Ensure ciphertext length is a multiple of AES block size
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	// Create CBC decrypter
	mode := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the ciphertext
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS#7 padding
	plaintext, err = pkcs7.Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("unpadding failed: %v", err)
	}

	// Convert from UTF-16LE to string
	password := utf16.DecodeUTF16LE(plaintext)

	return password, nil
}

// GPPPEncrypt encrypts a plaintext password using the Group Policy Preferences Password encryption algorithm
// It returns the base64-encoded encrypted string
func GPPPEncrypt(plaintext string) (string, error) {
	// Convert plaintext to UTF-16LE bytes
	plaintextBytes := utf16.EncodeUTF16LE(plaintext)

	// Apply PKCS#7 padding
	paddedPlaintext, err := pkcs7.Pad(plaintextBytes, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("failed to pad plaintext: %v", err)
	}

	// Fixed null IV (Initialization Vector)
	iv := make([]byte, aes.BlockSize)

	// Create AES cipher block
	block, err := aes.NewCipher(GPPP_AES_KEY)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// Create CBC encrypter
	mode := cipher.NewCBCEncrypter(block, iv)

	// Encrypt the plaintext
	ciphertext := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return encoded, nil
}
