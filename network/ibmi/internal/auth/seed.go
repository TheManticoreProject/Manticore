package auth

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// SeedLength is the size of a client or server seed.
const SeedLength = 8

// maxSeed is the largest seed the host servers accept; zero is also invalid.
const maxSeed = 0xDFFFFFFFFFFFFFFF

// maxSeedDraws bounds rejection sampling. Each draw is accepted with
// probability 7/8, so failing 64 times means the source is broken.
const maxSeedDraws = 64

// ErrSeedSource is returned when the random source cannot produce a seed.
var ErrSeedSource = errors.New("auth: cannot generate a client seed")

// NewSeed draws a client seed uniformly from [1, 0xDFFFFFFFFFFFFFFF] by
// rejection sampling. A nil r uses crypto/rand.
func NewSeed(r io.Reader) ([SeedLength]byte, error) {
	if r == nil {
		r = rand.Reader
	}
	var seed [SeedLength]byte
	for i := 0; i < maxSeedDraws; i++ {
		if _, err := io.ReadFull(r, seed[:]); err != nil {
			return [SeedLength]byte{}, fmt.Errorf("%w: %w", ErrSeedSource, err)
		}
		if ValidSeed(seed) {
			return seed, nil
		}
	}
	return [SeedLength]byte{}, fmt.Errorf("%w: no value in range after %d draws", ErrSeedSource, maxSeedDraws)
}

// ValidSeed reports whether seed is in the range the host servers accept.
func ValidSeed(seed [SeedLength]byte) bool {
	v := binary.BigEndian.Uint64(seed[:])
	return v != 0 && v <= maxSeed
}
