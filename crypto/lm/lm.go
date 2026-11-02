package lm

import (
	"crypto/des"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/unicode/norm"
)

// DefaultOEMCodePage is the OEM code page LMHash converts passwords to: CP437,
// the OEM code page of US-English Windows. Set it only during initialization:
// LMHash reads it on every call.
var DefaultOEMCodePage = charmap.CodePage437

// ErrUnmappableCharacter reports a password character the OEM code page cannot
// represent. Windows then computes no LM hash (RtlUpcaseUnicodeStringToOemString
// fails with STATUS_UNMAPPABLE_CHARACTER).
var ErrUnmappableCharacter = errors.New("lm: password character not representable in the OEM code page")

// emptyLMHash is the LM hash of the empty password (aad3b435b51404ee twice), the
// value tools show when an account has no LM hash.
var emptyLMHash = [16]byte{0xaa, 0xd3, 0xb4, 0x35, 0xb5, 0x14, 0x04, 0xee, 0xaa, 0xd3, 0xb4, 0x35, 0xb5, 0x14, 0x04, 0xee}

// LMHash computes the LM hash (LMOWFv1) of password as Windows does with the
// default OEM code page (CP437): see LMHashOEM. A password Windows cannot convert
// to the code page has no LM hash, and LMHash then returns the empty-password
// hash aad3b435b51404eeaad3b435b51404ee; use LMHashOEM to get the error.
func LMHash(password string) [16]byte {
	h, err := LMHashOEM(password, DefaultOEMCodePage)
	if err != nil {
		return emptyLMHash
	}
	return h
}

// LMHashOEM computes the LM hash of password in the given OEM code page, as
// Windows does: RtlUpcaseUnicodeStringToOemString upper-cases the password and
// converts it to the code page, and the first 14 bytes, NUL-padded, key the two
// DES halves that encrypt "KGS!@#$%".
//
// Upper-casing follows Windows' case table, which only maps a character with a
// true case pair: ÿ becomes Ÿ, but µ, ſ, ς and ı are left as they are, where
// Unicode would map them to Μ, S, Σ and I.
//
// A character the code page lacks is written as its best-fit equivalent, as
// Windows' best-fit tables do: an accented letter whose base letter the code
// page has (È becomes E in CP437), or one of the letters in bestFitLetters (ı
// and Ø become I and O). Any other unmappable character returns
// ErrUnmappableCharacter. This approximates Windows' best-fit tables for
// letters; their mappings of punctuation and symbols are not reproduced.
func LMHashOEM(password string, codePage *charmap.Charmap) ([16]byte, error) {
	oem, err := toUpperOEM(password, codePage)
	if err != nil {
		return [16]byte{}, err
	}
	return lmHashOfOEM(oem), nil
}

// toUpperOEM upper-cases password and converts it to the OEM code page, mapping
// an unrepresentable accented letter to its base letter.
func toUpperOEM(password string, codePage *charmap.Charmap) ([]byte, error) {
	if codePage == nil {
		codePage = DefaultOEMCodePage
	}
	out := make([]byte, 0, len(password))
	for _, r := range password {
		r = windowsUpper(r)
		if b, ok := codePage.EncodeRune(r); ok {
			out = append(out, b)
			continue
		}
		// Best fit: a letter Windows maps explicitly, or the base letter of a
		// decomposed accented letter.
		if fit, ok := bestFitLetters[r]; ok {
			if b, ok := codePage.EncodeRune(fit); ok {
				out = append(out, b)
				continue
			}
		}
		base := []rune(norm.NFD.String(string(r)))
		if len(base) > 1 {
			if b, ok := codePage.EncodeRune(base[0]); ok && base[0] < utf8.RuneSelf {
				out = append(out, b)
				continue
			}
		}
		return nil, fmt.Errorf("%w: %q", ErrUnmappableCharacter, r)
	}
	return out, nil
}

// windowsUpper upper-cases r as Windows' case table does: only when r has a true
// upper-case pair, i.e. lower-casing the upper-case form gives r back. Unicode
// maps µ, ſ, ς and ı to Μ, S, Σ and I, whose lower-case forms are other
// characters, and Windows leaves those four unchanged.
func windowsUpper(r rune) rune {
	u := unicode.ToUpper(r)
	if u != r && unicode.ToLower(u) != r {
		return r
	}
	return u
}

// bestFitLetters are best-fit mappings Windows applies to letters without a
// Unicode decomposition, as observed with RtlUpcaseUnicodeStringToOemString on
// code page 437.
var bestFitLetters = map[rune]rune{
	'ı': 'I',
	'Ø': 'O',
}

// lmHashOfOEM computes the LM one-way function of OEM password bytes: they are
// truncated or NUL-padded to 14 bytes and split into two DES keys.
func lmHashOfOEM(oem []byte) [16]byte {
	password := make([]byte, 14)
	copy(password, oem)

	// Split into two 7-byte halves
	firstHalf := []byte(password[:7])
	secondHalf := []byte(password[7:])

	// Convert 7-byte halves to 8-byte DES keys (with parity bits)
	firstKey := make([]byte, 8)
	secondKey := make([]byte, 8)

	// Add parity bits
	firstKey[0] = firstHalf[0]
	firstKey[1] = (firstHalf[0] << 7) | (firstHalf[1] >> 1)
	firstKey[2] = (firstHalf[1] << 6) | (firstHalf[2] >> 2)
	firstKey[3] = (firstHalf[2] << 5) | (firstHalf[3] >> 3)
	firstKey[4] = (firstHalf[3] << 4) | (firstHalf[4] >> 4)
	firstKey[5] = (firstHalf[4] << 3) | (firstHalf[5] >> 5)
	firstKey[6] = (firstHalf[5] << 2) | (firstHalf[6] >> 6)
	firstKey[7] = (firstHalf[6] << 1)

	secondKey[0] = secondHalf[0]
	secondKey[1] = (secondHalf[0] << 7) | (secondHalf[1] >> 1)
	secondKey[2] = (secondHalf[1] << 6) | (secondHalf[2] >> 2)
	secondKey[3] = (secondHalf[2] << 5) | (secondHalf[3] >> 3)
	secondKey[4] = (secondHalf[3] << 4) | (secondHalf[4] >> 4)
	secondKey[5] = (secondHalf[4] << 3) | (secondHalf[5] >> 5)
	secondKey[6] = (secondHalf[5] << 2) | (secondHalf[6] >> 6)
	secondKey[7] = (secondHalf[6] << 1)

	// The constant "KGS!@#$%" to encrypt
	magic := []byte("KGS!@#$%")

	// Create DES ciphers
	cipher1, _ := des.NewCipher(firstKey)
	cipher2, _ := des.NewCipher(secondKey)

	// Encrypt the magic constant with both keys
	result1 := make([]byte, 8)
	result2 := make([]byte, 8)
	cipher1.Encrypt(result1, magic)
	cipher2.Encrypt(result2, magic)

	// Concatenate results
	finalHash := append(result1, result2...)

	// Convert to lowercase hex string
	return [16]byte(finalHash)
}

// LMHashToHex converts the LM hash to a lowercase hex string
//
// # The LM hash is converted to a hex string by encoding the hash as a lowercase hex string
//
// Returns the lowercase hex string representation of the LM hash
func LMHashToHex(password string) string {
	lmHash := LMHash(password)
	return strings.ToLower(hex.EncodeToString(lmHash[:]))
}

// LMOWFv1, this is the same as LMHash
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/464551a8-9fc4-428e-b3d3-bc5bfb2e73a5
func LMOWFv1(Passwd, User, UserDom string) [16]byte {
	return LMHash(Passwd)
}
