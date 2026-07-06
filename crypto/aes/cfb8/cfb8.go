// Package cfb8 implements the 8-bit cipher feedback mode (CFB8) of operation for
// any block cipher, as defined by NIST SP 800-38A. Go's standard library exposes
// only full-block CFB (and deprecates it), whereas several Microsoft protocols —
// notably the Netlogon secure channel of MS-NRPC ([MS-NRPC] 3.1.4.4.1) — rely on
// the 8-bit-segment variant, in which the shift register advances one octet per
// plaintext octet.
//
// The implementations satisfy cipher.Stream, so they compose with any
// cipher.Block (AES, DES, …) exactly like the other stream modes in
// crypto/cipher. They are validated against the NIST SP 800-38A F.3.7/F.3.8
// (CFB8-AES128) known-answer vectors.
package cfb8

import "crypto/cipher"

// cfb8 is a CFB8 keystream generator over a block cipher. It maintains the shift
// register (initialised to the IV) and feeds back one octet of the segment per
// processed byte: the ciphertext octet, which is the output when encrypting and
// the input when decrypting.
type cfb8 struct {
	block     cipher.Block
	blockSize int
	shift     []byte
	tmp       []byte
	decrypt   bool
}

func newCFB8(block cipher.Block, iv []byte, decrypt bool) *cfb8 {
	bs := block.BlockSize()
	if len(iv) != bs {
		panic("cfb8: IV length must equal the block size")
	}
	shift := make([]byte, bs)
	copy(shift, iv)
	return &cfb8{
		block:     block,
		blockSize: bs,
		shift:     shift,
		tmp:       make([]byte, bs),
		decrypt:   decrypt,
	}
}

// NewEncrypter returns a cipher.Stream that encrypts with the given block cipher
// in CFB8 mode using the supplied IV. The IV length must equal the cipher's
// block size, otherwise NewEncrypter panics.
//
// Parameters:
//   - block: The block cipher to run in CFB8 mode (e.g. an AES cipher.Block).
//   - iv: The initialisation vector; its length must equal block.BlockSize().
//
// Returns:
//   - cipher.Stream: A stream that encrypts via XORKeyStream.
func NewEncrypter(block cipher.Block, iv []byte) cipher.Stream {
	return newCFB8(block, iv, false)
}

// NewDecrypter returns a cipher.Stream that decrypts with the given block cipher
// in CFB8 mode using the supplied IV. The IV length must equal the cipher's
// block size, otherwise NewDecrypter panics.
//
// Parameters:
//   - block: The block cipher to run in CFB8 mode (e.g. an AES cipher.Block).
//   - iv: The initialisation vector; its length must equal block.BlockSize().
//
// Returns:
//   - cipher.Stream: A stream that decrypts via XORKeyStream.
func NewDecrypter(block cipher.Block, iv []byte) cipher.Stream {
	return newCFB8(block, iv, true)
}

// XORKeyStream processes src into dst one octet at a time. For each octet it
// encrypts the shift register, XORs the high keystream octet with the input
// octet, and shifts the ciphertext octet into the register. dst may overlap src
// exactly (in-place) but must be at least len(src) long, otherwise XORKeyStream
// panics.
//
// Parameters:
//   - dst: The destination buffer (len(dst) >= len(src)); may alias src.
//   - src: The input octets to transform.
func (c *cfb8) XORKeyStream(dst, src []byte) {
	if len(dst) < len(src) {
		panic("cfb8: output smaller than input")
	}
	for i := range src {
		c.block.Encrypt(c.tmp, c.shift)
		in := src[i]
		out := in ^ c.tmp[0]
		dst[i] = out
		// Feed the ciphertext octet back into the register: that is the produced
		// octet when encrypting and the consumed octet when decrypting.
		feedback := out
		if c.decrypt {
			feedback = in
		}
		copy(c.shift, c.shift[1:])
		c.shift[c.blockSize-1] = feedback
	}
}
