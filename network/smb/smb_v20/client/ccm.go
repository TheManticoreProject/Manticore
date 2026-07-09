package client

import (
	"crypto/cipher"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
)

// AES-CCM (Counter with CBC-MAC), as specified in RFC 3610, implemented on top
// of a crypto/cipher.Block so no third-party dependency is required. SMB 3.x
// uses AES-128-CCM (and AES-256-CCM) as one of its two authenticated-encryption
// ciphers (MS-SMB2 3.1.4.3), with an 11-byte nonce and a 16-byte
// authentication tag.
//
// The interface mirrors crypto/cipher AEAD semantics: ccmSeal returns
// ciphertext || tag, and ccmOpen consumes ciphertext || tag and returns the
// plaintext, so both fit the same TRANSFORM_HEADER wrapping used by AES-GCM.

const ccmTagSize = 16

// ccmSeal encrypts and authenticates plaintext with associated data, returning
// ciphertext with the 16-byte tag appended. The nonce length determines the
// CCM length field L (L = 15 - len(nonce)); SMB uses an 11-byte nonce (L = 4).
func ccmSeal(block cipher.Block, nonce, plaintext, aad []byte) ([]byte, error) {
	if block.BlockSize() != 16 {
		return nil, fmt.Errorf("ccm: cipher block size must be 16, got %d", block.BlockSize())
	}
	l := 15 - len(nonce)
	if l < 2 || l > 8 {
		return nil, fmt.Errorf("ccm: invalid nonce length %d", len(nonce))
	}

	// The message length must fit in the L-byte length field.
	if l < 8 {
		maxLen := uint64(1) << (8 * uint(l))
		if uint64(len(plaintext)) >= maxLen {
			return nil, fmt.Errorf("ccm: plaintext too long for nonce length %d", len(nonce))
		}
	}

	tag := ccmTag(block, nonce, plaintext, aad, l, ccmTagSize)

	out := make([]byte, len(plaintext)+ccmTagSize)
	ccmCTR(block, nonce, l, plaintext, out[:len(plaintext)])

	// Encrypt the tag with counter block 0 (S_0) and append it.
	var counter0 [16]byte
	ccmCounterBlock(counter0[:], nonce, l, 0)
	var s0 [16]byte
	block.Encrypt(s0[:], counter0[:])
	for i := 0; i < ccmTagSize; i++ {
		out[len(plaintext)+i] = tag[i] ^ s0[i]
	}
	return out, nil
}

// ccmOpen verifies and decrypts ciphertext || tag with associated data,
// returning the plaintext. It returns an error if authentication fails.
func ccmOpen(block cipher.Block, nonce, ciphertextAndTag, aad []byte) ([]byte, error) {
	if block.BlockSize() != 16 {
		return nil, fmt.Errorf("ccm: cipher block size must be 16, got %d", block.BlockSize())
	}
	l := 15 - len(nonce)
	if l < 2 || l > 8 {
		return nil, fmt.Errorf("ccm: invalid nonce length %d", len(nonce))
	}
	if len(ciphertextAndTag) < ccmTagSize {
		return nil, fmt.Errorf("ccm: message shorter than tag")
	}

	ctLen := len(ciphertextAndTag) - ccmTagSize
	ciphertext := ciphertextAndTag[:ctLen]
	receivedTag := ciphertextAndTag[ctLen:]

	plaintext := make([]byte, ctLen)
	ccmCTR(block, nonce, l, ciphertext, plaintext)

	tag := ccmTag(block, nonce, plaintext, aad, l, ccmTagSize)
	var counter0 [16]byte
	ccmCounterBlock(counter0[:], nonce, l, 0)
	var s0 [16]byte
	block.Encrypt(s0[:], counter0[:])
	expectedTag := make([]byte, ccmTagSize)
	for i := 0; i < ccmTagSize; i++ {
		expectedTag[i] = tag[i] ^ s0[i]
	}

	if subtle.ConstantTimeCompare(expectedTag, receivedTag) != 1 {
		return nil, fmt.Errorf("ccm: authentication failed")
	}
	return plaintext, nil
}

// ccmTag computes the CBC-MAC T over the formatted blocks B_0, the AAD, and the
// plaintext (RFC 3610 section 2.2), returning the full-block MAC before it is
// encrypted with S_0. tagLen selects the M parameter encoded in the B_0 flags.
func ccmTag(block cipher.Block, nonce, plaintext, aad []byte, l, tagLen int) []byte {
	var b0 [16]byte
	// Flags: Adata bit (bit 6) | ((M-2)/2 << 3) | (L-1).
	flags := byte(l - 1)
	flags |= byte((tagLen-2)/2) << 3
	if len(aad) > 0 {
		flags |= 1 << 6
	}
	b0[0] = flags
	copy(b0[1:1+len(nonce)], nonce)
	// Message length in the trailing L bytes, big-endian.
	putUintBE(b0[16-l:], uint64(len(plaintext)), l)

	var mac [16]byte
	block.Encrypt(mac[:], b0[:])

	// Associated data, prefixed by its encoded length, zero-padded to a block.
	if len(aad) > 0 {
		var enc []byte
		alen := len(aad)
		switch {
		case alen < (1<<16 - 1<<8):
			enc = make([]byte, 2)
			binary.BigEndian.PutUint16(enc, uint16(alen))
		case uint64(alen) <= 0xFFFFFFFF:
			enc = make([]byte, 6)
			enc[0], enc[1] = 0xFF, 0xFE
			binary.BigEndian.PutUint32(enc[2:], uint32(alen))
		default:
			enc = make([]byte, 10)
			enc[0], enc[1] = 0xFF, 0xFF
			binary.BigEndian.PutUint64(enc[2:], uint64(alen))
		}
		// The encoded length and the associated data are concatenated and
		// zero-padded to a block boundary as a single unit before being folded in.
		adata := make([]byte, 0, len(enc)+len(aad))
		adata = append(adata, enc...)
		adata = append(adata, aad...)
		ccmCBCMAC(block, mac[:], adata)
	}

	ccmCBCMAC(block, mac[:], plaintext)
	return mac[:]
}

// ccmCBCMAC folds data into the running CBC-MAC state, zero-padding the final
// partial block, as required by CCM's B-block formatting.
func ccmCBCMAC(block cipher.Block, mac, data []byte) {
	var buf [16]byte
	for len(data) >= 16 {
		for i := 0; i < 16; i++ {
			mac[i] ^= data[i]
		}
		block.Encrypt(mac, mac)
		data = data[16:]
	}
	if len(data) > 0 {
		for i := range buf {
			buf[i] = 0
		}
		copy(buf[:], data)
		for i := 0; i < 16; i++ {
			mac[i] ^= buf[i]
		}
		block.Encrypt(mac, mac)
	}
}

// ccmCTR applies CCM counter-mode encryption/decryption to src, writing into
// dst. Counter blocks A_i (i >= 1) key the keystream; A_0 is reserved for the
// tag.
func ccmCTR(block cipher.Block, nonce []byte, l int, src, dst []byte) {
	var counter [16]byte
	var keystream [16]byte
	i := uint64(1)
	for offset := 0; offset < len(src); offset += 16 {
		ccmCounterBlock(counter[:], nonce, l, i)
		block.Encrypt(keystream[:], counter[:])
		end := offset + 16
		if end > len(src) {
			end = len(src)
		}
		for j := offset; j < end; j++ {
			dst[j] = src[j] ^ keystream[j-offset]
		}
		i++
	}
}

// ccmCounterBlock formats the CCM counter block A_i: a flags byte carrying
// (L-1), the nonce, and the L-byte counter value.
func ccmCounterBlock(dst, nonce []byte, l int, i uint64) {
	dst[0] = byte(l - 1)
	copy(dst[1:1+len(nonce)], nonce)
	putUintBE(dst[16-l:], i, l)
}

// putUintBE writes the low n bytes of v to dst in big-endian order.
func putUintBE(dst []byte, v uint64, n int) {
	for i := 0; i < n; i++ {
		dst[n-1-i] = byte(v >> (8 * uint(i)))
	}
}
