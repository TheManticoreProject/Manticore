// Package aescts implements AES-CTS (Ciphertext Stealing) mode as used by
// Kerberos per RFC 3962. The variant used is CBC-CTS where the last two
// ciphertext blocks are swapped before output (Kerberos / CS3 style).
package cts

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

const blockSize = 16

// Encrypt encrypts plaintext using AES-CTS with the given key and IV.
// plaintext must be >= 16 bytes (one AES block).
// Output length equals input length.
//
// For plaintext of exactly one block (16 bytes), standard AES-CBC is used
// (no swap is possible with a single block). For two or more blocks, the
// last two blocks in the CBC output are swapped and the output is truncated
// to len(plaintext) bytes.
func Encrypt(key, iv, plaintext []byte) ([]byte, error) {
	n := len(plaintext)
	if n < blockSize {
		return nil, errors.New("aescts: plaintext must be at least 16 bytes")
	}

	// Special case: exactly one block — CBC with no swap
	if n == blockSize {
		padded := make([]byte, blockSize)
		copy(padded, plaintext)
		return aesCBCEncrypt(key, iv, padded)
	}

	r := n % blockSize // remainder bytes in last partial block (0 = exact multiple)

	// Pad plaintext to a multiple of blockSize
	paddedLen := n
	if r != 0 {
		paddedLen = n + (blockSize - r)
	}
	padded := make([]byte, paddedLen)
	copy(padded, plaintext)

	// AES-CBC encrypt the padded plaintext
	cbcOut, err := aesCBCEncrypt(key, iv, padded)
	if err != nil {
		return nil, err
	}

	numBlocks := paddedLen / blockSize
	result := make([]byte, n)

	if r == 0 {
		// Exact multiple of blockSize: swap last two complete blocks.
		// CBC output: ... C[n-2] C[n-1]
		// CTS output: ... C[n-1] C[n-2]
		prefixEnd := n - 2*blockSize
		if prefixEnd > 0 {
			copy(result[:prefixEnd], cbcOut[:prefixEnd])
		}
		copy(result[prefixEnd:prefixEnd+blockSize], cbcOut[n-blockSize:n])
		copy(result[prefixEnd+blockSize:n], cbcOut[n-2*blockSize:n-blockSize])
	} else {
		// Non-multiple: CBC gives numBlocks full blocks (last one zero-padded).
		// CBC blocks: C[0] ... C[numBlocks-2] C[numBlocks-1]
		// CTS output: C[0]...C[numBlocks-3] + C[numBlocks-1](full 16B) + C[numBlocks-2][:r]
		// Total: (numBlocks-2)*16 + 16 + r = (numBlocks-1)*16 + r = n ✓
		prefixEnd := (numBlocks - 2) * blockSize
		penultStart := (numBlocks - 2) * blockSize
		lastStart := (numBlocks - 1) * blockSize

		if prefixEnd > 0 {
			copy(result[:prefixEnd], cbcOut[:prefixEnd])
		}
		// Full last CBC block
		copy(result[prefixEnd:prefixEnd+blockSize], cbcOut[lastStart:lastStart+blockSize])
		// First r bytes of penultimate CBC block
		copy(result[prefixEnd+blockSize:n], cbcOut[penultStart:penultStart+r])
	}

	return result, nil
}

// Decrypt decrypts ciphertext using AES-CTS with the given key and IV.
// ciphertext must be >= 16 bytes.
// Output length equals input length.
func Decrypt(key, iv, ciphertext []byte) ([]byte, error) {
	n := len(ciphertext)
	if n < blockSize {
		return nil, errors.New("aescts: ciphertext must be at least 16 bytes")
	}

	// Special case: exactly one block — CBC with no un-swap
	if n == blockSize {
		return aesCBCDecrypt(key, iv, ciphertext)
	}

	r := n % blockSize

	if r == 0 {
		// Un-swap last two full blocks, then normal CBC decrypt.
		buf := make([]byte, n)
		copy(buf[:n-2*blockSize], ciphertext[:n-2*blockSize])
		copy(buf[n-2*blockSize:n-blockSize], ciphertext[n-blockSize:n])
		copy(buf[n-blockSize:n], ciphertext[n-2*blockSize:n-blockSize])
		return aesCBCDecrypt(key, iv, buf)
	}

	// Non-multiple case.
	// CTS ciphertext layout (from Encrypt):
	//   prefix: (numBlocks-2)*16 bytes  — blocks C[0]..C[numBlocks-3]
	//   Clast:  16 bytes                — last CBC block (C[numBlocks-1])
	//   Cpen_partial: r bytes           — first r bytes of penultimate CBC block (C[numBlocks-2])
	//
	// To reconstruct the penultimate CBC block (C[numBlocks-2]):
	//   AES-ECB-decrypt Clast → X  (= P[numBlocks-1]_padded XOR C[numBlocks-2])
	//   Since P[numBlocks-1] was zero-padded, X[r:] = 0 XOR C[numBlocks-2][r:] = C[numBlocks-2][r:]
	//   So full C[numBlocks-2] = Cpen_partial + X[r:]
	//
	// Recover last r bytes of plaintext:
	//   P[numBlocks-1][:r] = X[:r] XOR C[numBlocks-2][:r] = X[:r] XOR Cpen_partial
	//
	// Decrypt prefix + C[numBlocks-2] with CBC to get P[0]..P[numBlocks-2].

	numBlocks := n / blockSize // integer division; excludes the partial block
	prefixLen := (numBlocks - 1) * blockSize
	clast := ciphertext[prefixLen : prefixLen+blockSize]
	cpenPartial := ciphertext[prefixLen+blockSize:]

	// AES-ECB decrypt Clast
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	x := make([]byte, blockSize)
	blockCipher.Decrypt(x, clast)

	// Reconstruct full penultimate CBC block
	cpen := make([]byte, blockSize)
	copy(cpen[:r], cpenPartial)
	copy(cpen[r:], x[r:])

	// Recover last r bytes of plaintext
	lastPartial := make([]byte, r)
	for i := 0; i < r; i++ {
		lastPartial[i] = x[i] ^ cpenPartial[i]
	}

	// CBC-decrypt prefix + cpen to get P[0]..P[numBlocks-2]
	cbcInput := make([]byte, prefixLen+blockSize)
	copy(cbcInput[:prefixLen], ciphertext[:prefixLen])
	copy(cbcInput[prefixLen:], cpen)

	mainPlain, err := aesCBCDecrypt(key, iv, cbcInput)
	if err != nil {
		return nil, err
	}

	result := append(mainPlain, lastPartial...)
	return result, nil
}

// aesCBCEncrypt performs standard AES-CBC encryption.
// plaintext length must be a multiple of blockSize.
func aesCBCEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(blockCipher, iv)
	mode.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

// aesCBCDecrypt performs standard AES-CBC decryption.
// ciphertext length must be a multiple of blockSize.
func aesCBCDecrypt(key, iv, ciphertext []byte) ([]byte, error) {
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(blockCipher, iv)
	mode.CryptBlocks(plaintext, ciphertext)
	return plaintext, nil
}
