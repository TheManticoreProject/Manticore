package md4

import (
	"encoding/binary"
	"encoding/hex"
)

// Source: https://dspace.mit.edu/bitstream/handle/1721.1/149165/MIT-LCS-TM-434.pdf

const (
	chunkSize = 64
	init0     = 0x67452301
	init1     = 0xefcdab89
	init2     = 0x98badcfe
	init3     = 0x10325476
)

// MD4 is a struct that implements the MD4 hash algorithm.
//
// Fields:
// - state: An array of four 32-bit unsigned integers that represent the internal state of the MD4 hash.
// - count: A uint64 that represents the number of bits processed so far.
// - buffer: An array of 64 bytes that is used to store the data to be hashed.
type MD4 struct {
	state  [4]uint32
	count  uint64
	buffer [chunkSize]byte
}

func New() *MD4 {
	md4 := &MD4{}
	md4.state[0] = init0
	md4.state[1] = init1
	md4.state[2] = init2
	md4.state[3] = init3
	return md4
}

// Write adds more data to the running MD4 hash. It can be called multiple times
// to add more data. The MD4 hash is computed incrementally as data is written.
//
// Parameters:
//   - p: A byte slice containing the data to be added to the hash.
//
// Returns:
//   - n: The number of bytes written from p.
//   - err: An error value, which is always nil in this implementation.
//
// The function updates the internal state of the MD4 object with the new data.
// It processes the data in chunks of 64 bytes (the MD4 block size). If there is
// any remaining data that is less than a full chunk, it is buffered for the next
// call to Write or for the final hash computation.
func (md4 *MD4) Write(p []byte) (n int, err error) {
	n = len(p)
	md4.count += uint64(n) * 8

	buffered := int((md4.count/8 - uint64(n)) % chunkSize)
	remaining := chunkSize - buffered

	var i int
	if n >= remaining {
		copy(md4.buffer[buffered:], p[:remaining])
		md4.processChunk(md4.buffer[:])
		i = remaining
		for ; i+chunkSize <= n; i += chunkSize {
			md4.processChunk(p[i : i+chunkSize])
		}
		buffered = 0
	}
	copy(md4.buffer[buffered:], p[i:])
	return
}

// Sum returns the MD4 checksum of the data.
//
// Parameters:
//   - data: A byte slice containing the data to be hashed.
//
// Returns:
//   - A 16-byte array containing the MD4 checksum of the input data.
//
// The function creates a new MD4 hash object, writes the input data to it, and
// then computes the final hash value. The resulting 16-byte checksum is returned
// as an array.
func (md4 *MD4) Sum() [16]byte {
	var padding [64]byte
	padding[0] = 0x80
	bits := [8]byte{}
	binary.LittleEndian.PutUint64(bits[:], md4.count)
	index := md4.count / 8 % chunkSize
	padLen := 56 - index
	if index >= 56 {
		padLen += chunkSize
	}
	md4.Write(padding[:padLen])
	md4.Write(bits[:])
	var digest [16]byte
	for i, s := range md4.state {
		binary.LittleEndian.PutUint32(digest[i*4:], s)
	}
	return digest
}

// Sum returns the MD4 checksum of the given data.
//
// Parameters:
//   - data: A byte slice containing the data to be hashed.
//
// Returns:
//   - A 16-byte array containing the MD4 checksum of the input data.
//
// The function creates a new MD4 hash object, writes the input data to it, and
// then computes the final hash value. The resulting 16-byte checksum is returned
// as an array.
func Sum(data []byte) [16]byte {
	h := New()
	h.Write(data)
	return h.Sum()
}

// HexSum returns the MD4 checksum of the data as a hexadecimal string.
//
// Returns:
//   - A string containing the MD4 checksum of the input data in hexadecimal format.
//
// The function computes the MD4 checksum of the data and encodes it as a
// hexadecimal string. This is useful for representing the hash value in a
// human-readable format.
func (md4 *MD4) HexSum() string {
	digest := md4.Sum()
	return hex.EncodeToString(digest[:])
}

// processChunk processes a single 512-bit chunk of the input data.
//
// Parameters:
//   - chunk: A byte slice containing the 512-bit chunk to be processed.
//
// The function updates the MD4 state by processing the given chunk using the MD4
// compression function. It performs three rounds of operations on the chunk,
// using the auxiliary functions ff, gg, and hh to transform the state.
//
// The chunk is divided into sixteen 32-bit words, which are used as input to the
// compression function. The state is updated in-place, and the function does not
// return any value.
func (md4 *MD4) processChunk(chunk []byte) {
	var x [16]uint32
	for i := 0; i < 16; i++ {
		x[i] = binary.LittleEndian.Uint32(chunk[i*4:])
	}
	a, b, c, d := md4.state[0], md4.state[1], md4.state[2], md4.state[3]

	// Round 1
	a = ff(a, b, c, d, x[0], 3)
	d = ff(d, a, b, c, x[1], 7)
	c = ff(c, d, a, b, x[2], 11)
	b = ff(b, c, d, a, x[3], 19)
	a = ff(a, b, c, d, x[4], 3)
	d = ff(d, a, b, c, x[5], 7)
	c = ff(c, d, a, b, x[6], 11)
	b = ff(b, c, d, a, x[7], 19)
	a = ff(a, b, c, d, x[8], 3)
	d = ff(d, a, b, c, x[9], 7)
	c = ff(c, d, a, b, x[10], 11)
	b = ff(b, c, d, a, x[11], 19)
	a = ff(a, b, c, d, x[12], 3)
	d = ff(d, a, b, c, x[13], 7)
	c = ff(c, d, a, b, x[14], 11)
	b = ff(b, c, d, a, x[15], 19)

	// Round 2
	a = gg(a, b, c, d, x[0], 3)
	d = gg(d, a, b, c, x[4], 5)
	c = gg(c, d, a, b, x[8], 9)
	b = gg(b, c, d, a, x[12], 13)
	a = gg(a, b, c, d, x[1], 3)
	d = gg(d, a, b, c, x[5], 5)
	c = gg(c, d, a, b, x[9], 9)
	b = gg(b, c, d, a, x[13], 13)
	a = gg(a, b, c, d, x[2], 3)
	d = gg(d, a, b, c, x[6], 5)
	c = gg(c, d, a, b, x[10], 9)
	b = gg(b, c, d, a, x[14], 13)
	a = gg(a, b, c, d, x[3], 3)
	d = gg(d, a, b, c, x[7], 5)
	c = gg(c, d, a, b, x[11], 9)
	b = gg(b, c, d, a, x[15], 13)

	// Round 3
	a = hh(a, b, c, d, x[0], 3)
	d = hh(d, a, b, c, x[8], 9)
	c = hh(c, d, a, b, x[4], 11)
	b = hh(b, c, d, a, x[12], 15)
	a = hh(a, b, c, d, x[2], 3)
	d = hh(d, a, b, c, x[10], 9)
	c = hh(c, d, a, b, x[6], 11)
	b = hh(b, c, d, a, x[14], 15)
	a = hh(a, b, c, d, x[1], 3)
	d = hh(d, a, b, c, x[9], 9)
	c = hh(c, d, a, b, x[5], 11)
	b = hh(b, c, d, a, x[13], 15)
	a = hh(a, b, c, d, x[3], 3)
	d = hh(d, a, b, c, x[11], 9)
	c = hh(c, d, a, b, x[7], 11)
	b = hh(b, c, d, a, x[15], 15)

	md4.state[0] += a
	md4.state[1] += b
	md4.state[2] += c
	md4.state[3] += d
}

func ff(a, b, c, d, x, s uint32) uint32 {
	return rol(a+(d^(b&(c^d)))+x, s)
}

func gg(a, b, c, d, x, s uint32) uint32 {
	return rol(a+((b&c)|(d&(b|c)))+x+0x5a827999, s)
}

func hh(a, b, c, d, x, s uint32) uint32 {
	return rol(a+(b^c^d)+x+0x6ed9eba1, s)
}

func rol(x, s uint32) uint32 {
	return (x << s) | (x >> (32 - s))
}
