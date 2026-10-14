package auth

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf16"
)

// ProfileLength is the fixed width of a sign-on user profile.
const ProfileLength = 10

const ebcdicBlank = 0x40

// ErrInvalidProfile is matched (via errors.Is) by profile encoding failures.
// Errors report positions, never the characters themselves.
var ErrInvalidProfile = errors.New("auth: invalid user profile")

// nationalAliases are the characters the sign-on table collapses onto the
// three variant characters.
var nationalAliases = map[rune]rune{
	'£': '#', 'Ä': '#', 'Æ': '#', 'Ñ': '#',
	'¥': '$', 'Å': '$', 'İ': '$',
	'§': '@', 'Ð': '@', 'Ö': '@', 'Ø': '@', 'à': '@', 'Ş': '@',
}

// mapSignonRune maps one character through the closed sign-on table. It
// returns the EBCDIC byte and the canonical character, or ok=false. A space
// is not a member: padding is added by the encoders, never accepted as input.
func mapSignonRune(r rune) (b byte, canonical rune, ok bool) {
	if c, alias := nationalAliases[r]; alias {
		r = c
	}
	switch {
	case r == '#':
		return 0x7B, r, true
	case r == '$':
		return 0x5B, r, true
	case r == '@':
		return 0x7C, r, true
	case r == '_':
		return 0x6D, r, true
	case r >= '0' && r <= '9':
		return byte(0xF0 + r - '0'), r, true
	case r >= 'A' && r <= 'I':
		return byte(0xC1 + r - 'A'), r, true
	case r >= 'J' && r <= 'R':
		return byte(0xD1 + r - 'J'), r, true
	case r >= 'S' && r <= 'Z':
		return byte(0xE2 + r - 'S'), r, true
	}
	return 0, 0, false
}

// Profile is a user profile in its two sign-on encodings.
type Profile struct {
	// EBCDIC is the blank-padded sign-on encoding sent in CP 0x1104 and used
	// by the DES algorithm.
	EBCDIC [ProfileLength]byte
	// canonical is the profile after the table (aliases collapsed), space
	// padded; the SHA-1 algorithm hashes its UTF-16BE form.
	canonical [ProfileLength]rune
}

// EncodeProfile uppercases name rune by rune (locale-independent simple case
// mapping) and maps it through the sign-on table. It must be 1 to 10
// characters, all in the table.
func EncodeProfile(name string) (Profile, error) {
	var p Profile
	for i := range p.EBCDIC {
		p.EBCDIC[i] = ebcdicBlank
		p.canonical[i] = ' '
	}
	n := 0
	for _, r := range name {
		if n == ProfileLength {
			return Profile{}, fmt.Errorf("%w: longer than %d characters", ErrInvalidProfile, ProfileLength)
		}
		b, c, ok := mapSignonRune(unicode.ToUpper(r))
		if !ok {
			return Profile{}, fmt.Errorf("%w: character %d is outside the sign-on character set", ErrInvalidProfile, n+1)
		}
		p.EBCDIC[n], p.canonical[n] = b, c
		n++
	}
	if n == 0 {
		return Profile{}, fmt.Errorf("%w: empty", ErrInvalidProfile)
	}
	return p, nil
}

// UTF16BE returns the canonical profile as 20 bytes of UTF-16BE.
func (p Profile) UTF16BE() []byte {
	return utf16BE(p.canonical[:])
}

func utf16BE(runes []rune) []byte {
	units := utf16.Encode(runes)
	out := make([]byte, 2*len(units))
	for i, u := range units {
		out[2*i], out[2*i+1] = byte(u>>8), byte(u)
	}
	wipeUint16(units)
	return out
}
