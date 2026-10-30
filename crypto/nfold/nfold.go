// Package nfold implements the N-FOLD function from RFC 3961 Section 5.1.
// N-FOLD is used by Kerberos to generate key derivation constants.
package nfold

// gcd computes the greatest common divisor of a and b.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// lcm computes the least common multiple of a and b.
func lcm(a, b int) int {
	return a / gcd(a, b) * b
}

// getBit returns the value (0 or 1) of bit p in b.
// Bit 0 is the MSB of b[0], bit 7 is the LSB of b[0], bit 8 is MSB of b[1], etc.
func getBit(b []byte, p int) int {
	pByte := p / 8
	pBit := uint(p % 8)
	return int(b[pByte]>>(8-(pBit+1))) & 0x01
}

// setBit sets bit p in b to v (0 or 1).
func setBit(b []byte, p, v int) {
	pByte := p / 8
	pBit := uint(p % 8)
	b[pByte] = byte(v<<(8-(pBit+1))) | b[pByte]
}

// rotateRight performs a bit-level right rotation of b by step positions.
func rotateRight(b []byte, step int) []byte {
	out := make([]byte, len(b))
	bitLen := len(b) * 8
	for i := 0; i < bitLen; i++ {
		v := getBit(b, i)
		setBit(out, (i+step)%bitLen, v)
	}
	return out
}

// onesComplementAddition adds two equal-length byte slices using ones' complement
// (end-around carry) arithmetic, processing from LSB to MSB.
func onesComplementAddition(n1, n2 []byte) []byte {
	numBits := len(n1) * 8
	out := make([]byte, len(n1))
	carry := 0
	for i := numBits - 1; i >= 0; i-- {
		s := getBit(n1, i) + getBit(n2, i) + carry
		setBit(out, i, s&1)
		carry = s >> 1
	}
	if carry == 1 {
		// End-around carry: add 1 to the result
		carryBuf := make([]byte, len(n1))
		carryBuf[len(carryBuf)-1] = 1
		out = onesComplementAddition(out, carryBuf)
	}
	return out
}

// NFold folds the input byte string into n bits (n must be a multiple of 8).
//
// The algorithm (RFC 3961 Section 5.1):
//  1. Let k = len(in)*8 and l = lcm(n, k).
//  2. Build a buffer of l/8 bytes by concatenating l/k copies of the input,
//     each copy rotated right by 13*i bits relative to the original.
//  3. XOR (with end-around carry / ones' complement addition) all n-bit
//     blocks of the buffer together to produce the n-bit output.
//
// An empty input or a non-positive n has no blocks to fold; the result is then
// n/8 zero bytes (an empty slice for n <= 0) rather than a division by zero.
func NFold(in []byte, n int) []byte {
	if len(in) == 0 || n <= 0 {
		return make([]byte, max(n, 0)/8)
	}

	k := len(in) * 8
	lcmVal := lcm(n, k)
	numCopies := lcmVal / k

	// Build the concatenated rotated buffer
	var buf []byte
	for i := 0; i < numCopies; i++ {
		buf = append(buf, rotateRight(in, 13*i)...)
	}

	// Ones' complement addition of all n-bit (n/8 byte) blocks
	result := make([]byte, n/8)
	block := make([]byte, n/8)
	numBlocks := lcmVal / n
	for i := 0; i < numBlocks; i++ {
		copy(block, buf[i*(n/8):(i+1)*(n/8)])
		result = onesComplementAddition(result, block)
	}
	return result
}
