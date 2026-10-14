package auth

import (
	"crypto/des" // Required for password levels 0/1 wire compatibility.
	"encoding/binary"
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

// DESSubstituteLength is the size of a level 0/1 substitute.
const DESSubstituteLength = 8

const desPasswordLength = 10

// ErrInvalidPassword is matched (via errors.Is) when a password cannot be
// used with the negotiated password level. Errors never include password
// characters.
var ErrInvalidPassword = errors.New("auth: invalid password")

var sequenceOne = [8]byte{0, 0, 0, 0, 0, 0, 0, 1}

// encodeDESPassword uppercases the UTF-8 password, prefixes Q to an
// all-digit password, and maps it through the sign-on table into 10
// blank-padded bytes. The prefix is applied before the length check, so at
// most nine digits fit.
func encodeDESPassword(password []byte) ([desPasswordLength]byte, error) {
	var out [desPasswordLength]byte
	if len(password) == 0 {
		return out, fmt.Errorf("%w: empty", ErrInvalidPassword)
	}
	if !utf8.Valid(password) {
		return out, fmt.Errorf("%w: not valid UTF-8", ErrInvalidPassword)
	}
	runes := make([]rune, 0, len(password)+1)
	defer func() { wipeRunes(runes[:cap(runes)]) }()
	numeric := true
	for i := 0; i < len(password); {
		r, size := utf8.DecodeRune(password[i:])
		i += size
		r = unicode.ToUpper(r)
		if r < '0' || r > '9' {
			numeric = false
		}
		runes = append(runes, r)
	}
	if numeric {
		runes = append([]rune{'Q'}, runes...)
	}
	if len(runes) > desPasswordLength {
		return out, fmt.Errorf("%w: longer than %d characters for password level 0/1", ErrInvalidPassword, desPasswordLength)
	}
	for i := range out {
		out[i] = ebcdicBlank
	}
	for i, r := range runes {
		b, _, ok := mapSignonRune(r)
		if !ok {
			wipe(out[:])
			return out, fmt.Errorf("%w: character %d is outside the password level 0/1 character set", ErrInvalidPassword, i+1)
		}
		out[i] = b
	}
	return out, nil
}

// foldProfile returns the first eight profile bytes, folding bytes 8 and 9
// into them when the profile is longer than eight characters.
func foldProfile(u [ProfileLength]byte) [8]byte {
	var out [8]byte
	copy(out[:], u[:8])
	if u[8] == ebcdicBlank {
		return out // at most eight characters: no folding
	}
	out[0] ^= u[8] & 0xC0
	out[1] ^= (u[8] & 0x30) << 2
	out[2] ^= (u[8] & 0x0C) << 4
	out[3] ^= (u[8] & 0x03) << 6
	out[4] ^= u[9] & 0xC0
	out[5] ^= (u[9] & 0x30) << 2
	out[6] ^= (u[9] & 0x0C) << 4
	out[7] ^= (u[9] & 0x03) << 6
	return out
}

func desEncrypt(key, block [8]byte) [8]byte {
	c, err := des.NewCipher(key[:])
	if err != nil {
		panic("auth: DES key of fixed size rejected: " + err.Error()) // unreachable: key is 8 bytes
	}
	var out [8]byte
	c.Encrypt(out[:], block[:])
	return out
}

func xor8(a, b [8]byte) [8]byte {
	for i := range a {
		a[i] ^= b[i]
	}
	return a
}

// desPasswordToken derives the level 0/1 password token. Each 8-byte
// password block is XORed with 0x55 bytes and shifted left one bit as a
// 64-bit value, then used as the DES key to encrypt the folded profile. A
// 9- or 10-character password yields a second token from its tail, and the
// two are XORed.
func desPasswordToken(profile [ProfileLength]byte, password [desPasswordLength]byte) [8]byte {
	u := foldProfile(profile)
	tokenFor := func(block [8]byte) [8]byte {
		v := (binary.BigEndian.Uint64(block[:]) ^ 0x5555555555555555) << 1
		var key [8]byte
		binary.BigEndian.PutUint64(key[:], v)
		t := desEncrypt(key, u)
		wipe(key[:])
		return t
	}
	var first [8]byte
	copy(first[:], password[:8])
	token := tokenFor(first)
	wipe(first[:])
	if password[8] == ebcdicBlank && password[9] == ebcdicBlank {
		return token
	}
	second := [8]byte{password[8], password[9], ebcdicBlank, ebcdicBlank, ebcdicBlank, ebcdicBlank, ebcdicBlank, ebcdicBlank}
	t2 := tokenFor(second)
	wipe(second[:])
	token = xor8(token, t2)
	wipe(t2[:])
	return token
}

// desSubstitute mixes the token with both seeds, the profile and the
// sequence number into the 8-byte level 0/1 substitute.
func desSubstitute(token [8]byte, profile [ProfileLength]byte, clientSeed, serverSeed [SeedLength]byte) [DESSubstituteLength]byte {
	var rdrSeq [8]byte
	binary.BigEndian.PutUint64(rdrSeq[:], binary.BigEndian.Uint64(serverSeed[:])+1) // + sequence, mod 2^64

	e1 := desEncrypt(token, rdrSeq)
	verifier := desEncrypt(token, xor8(e1, clientSeed))
	var head [8]byte
	copy(head[:], profile[:8])
	e3 := desEncrypt(token, xor8(xor8(head, rdrSeq), verifier))
	tail := [8]byte{profile[8], profile[9], ebcdicBlank, ebcdicBlank, ebcdicBlank, ebcdicBlank, ebcdicBlank, ebcdicBlank}
	e4 := desEncrypt(token, xor8(xor8(tail, rdrSeq), e3))
	sub := desEncrypt(token, xor8(e4, sequenceOne))
	for _, b := range [][8]byte{e1, verifier, e3, e4} {
		wipe(b[:])
	}
	return sub
}
