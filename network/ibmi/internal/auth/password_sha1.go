package auth

import (
	"crypto/sha1" // Required for password levels 2/3 wire compatibility.
	"fmt"
	"unicode/utf8"
)

// SHA1SubstituteLength is the size of a level 2/3 substitute.
const SHA1SubstituteLength = 20

// maxSHAPasswordUnits is the longest accepted password, in UTF-16 code
// units, matching the IBM i 128-character password limit.
const maxSHAPasswordUnits = 128

// shaPasswordUTF16BE validates a level 2/3 password and returns it as
// UTF-16BE: a leading '*' is rejected, trailing NUL, space and ideographic
// space are removed, and an empty result is rejected.
func shaPasswordUTF16BE(password []byte) ([]byte, error) {
	if !utf8.Valid(password) {
		return nil, fmt.Errorf("%w: not valid UTF-8", ErrInvalidPassword)
	}
	runes := make([]rune, 0, utf8.RuneCount(password))
	defer func() { wipeRunes(runes[:cap(runes)]) }()
	for i := 0; i < len(password); {
		r, size := utf8.DecodeRune(password[i:])
		i += size
		runes = append(runes, r)
	}
	if len(runes) > 0 && runes[0] == '*' {
		return nil, fmt.Errorf("%w: a password may not start with '*'", ErrInvalidPassword)
	}
	end := len(runes)
	for end > 0 && (runes[end-1] == 0 || runes[end-1] == ' ' || runes[end-1] == '　') {
		end--
	}
	if end == 0 {
		return nil, fmt.Errorf("%w: empty after removing trailing blanks", ErrInvalidPassword)
	}
	out := utf16BE(runes[:end])
	if len(out)/2 > maxSHAPasswordUnits {
		wipe(out)
		return nil, fmt.Errorf("%w: longer than %d UTF-16 code units", ErrInvalidPassword, maxSHAPasswordUnits)
	}
	return out, nil
}

// sha1Token is SHA1(profile UTF-16BE || password UTF-16BE).
func sha1Token(profileUTF16, passwordUTF16 []byte) [sha1.Size]byte {
	h := sha1.New()
	h.Write(profileUTF16)
	h.Write(passwordUTF16)
	var out [sha1.Size]byte
	h.Sum(out[:0])
	return out
}

// sha1Substitute is SHA1(token || server seed || client seed || profile
// UTF-16BE || sequence).
func sha1Substitute(token [sha1.Size]byte, profileUTF16 []byte, clientSeed, serverSeed [SeedLength]byte) [SHA1SubstituteLength]byte {
	h := sha1.New()
	h.Write(token[:])
	h.Write(serverSeed[:])
	h.Write(clientSeed[:])
	h.Write(profileUTF16)
	h.Write(sequenceOne[:])
	var out [SHA1SubstituteLength]byte
	h.Sum(out[:0])
	return out
}
