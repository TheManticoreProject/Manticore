package rc4

import (
	"fmt"
	"unsafe"
)

// KeySizeError is returned when the provided key is not of the correct size.
type KeySizeError int

// Error implements the error interface.
// Parameters:
//   - k: The KeySizeError value representing the invalid key size
//
// Returns:
//   - string: A formatted error message indicating the invalid key size
func (k KeySizeError) Error() string {
	return fmt.Sprintf("rc4: invalid key size %d", k)
}

// RC4 represents an RC4 cipher instance.
type RC4 struct {
	// s is the S-box.
	s [256]uint8

	// i and j are the indices.
	i, j uint8

	// Key is the key used to initialize the cipher.
	Key []byte
}

// NewRC4 creates and returns a new RC4 cipher.
// The key is empty, so the cipher is not initialized.
// Parameters:
//   - None
//
// Returns:
//   - *RC4: A pointer to a new RC4 cipher instance
//   - error: An error if the cipher could not be created
func NewRC4() (*RC4, error) {
	return NewRC4WithKey([]byte{})
}

// NewRC4WithKey creates and returns a new RC4 cipher with the given key.
// The key must be between 1 and 256 bytes in length.
// Parameters:
//   - key: The byte slice containing the key for the RC4 cipher
//
// Returns:
//   - *RC4: A pointer to a new RC4 cipher instance initialized with the key
//   - error: An error if the key size is invalid
func NewRC4WithKey(key []byte) (*RC4, error) {
	k := len(key)
	if k < 1 || k > 256 {
		return nil, KeySizeError(k)
	}

	c := &RC4{
		Key: key,
	}

	// Initialize S-box
	for i := 0; i < 256; i++ {
		c.s[i] = uint8(i)
	}

	// Key scheduling algorithm
	var j uint8
	for i := 0; i < 256; i++ {
		j += c.s[i] + key[i%k]
		c.s[i], c.s[j] = c.s[j], c.s[i]
	}

	// Initialize indices
	c.i = 0
	c.j = 0

	return c, nil
}

// Reset zeros the key data and resets the cipher to its initial state.
// Parameters:
//   - c: The RC4 cipher instance to reset
//
// Returns:
//   - None
func (c *RC4) Reset() {
	// Reset key
	c.Key = []byte{}

	// Reset S-box
	for i := 0; i < 256; i++ {
		c.s[i] = uint8(i)
	}

	// Reset indices
	c.i = 0
	c.j = 0
}

// XORKeyStream applies the RC4 cipher to the byte slice.
// It encrypts or decrypts the bytes in src and writes the result to dst.
// Src and dst may be the same slice but otherwise should not overlap.
// Parameters:
//   - dst: The destination byte slice where the result will be written
//   - src: The source byte slice containing the data to be encrypted/decrypted
//
// Returns:
//   - None
//
// Panics:
//   - If dst is smaller than src
//   - If src and dst overlap but are not identical slices
func (c *RC4) XORKeyStream(dst, src []byte) {
	if len(dst) < len(src) {
		panic("rc4: output smaller than input")
	}

	// Only allow in-place (identical slice) or fully disjoint buffers.
	if len(src) > 0 && &dst[0] != &src[0] {
		// Compute address ranges
		d0 := uintptr(unsafe.Pointer(&dst[0]))
		dN := d0 + uintptr(len(dst)-1)
		s0 := uintptr(unsafe.Pointer(&src[0]))
		sN := s0 + uintptr(len(src)-1)

		// If ranges intersect, panic
		if d0 <= sN && s0 <= dN {
			panic("rc4: overlapping src and dst but not identical slices")
		}
	}

	// Pseudo-random generation algorithm (PRGA)
	i, j := c.i, c.j
	for k, v := range src {
		i++
		j += c.s[i]
		c.s[i], c.s[j] = c.s[j], c.s[i]
		t := uint8(int(c.s[i]) + int(c.s[j]))
		dst[k] = v ^ c.s[t]
	}
	c.i, c.j = i, j
}
