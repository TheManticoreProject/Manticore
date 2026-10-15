package auth

import (
	"crypto/sha256"
	"crypto/sha512" // Required for password level 4 wire compatibility.
	"unicode/utf16"

	"golang.org/x/crypto/pbkdf2"
)

// SHA512SubstituteLength is the size of a level 4 substitute.
const SHA512SubstituteLength = 64

// level4Iterations is the PBKDF2 iteration count of password level 4.
const level4Iterations = 10022

// level4Salt is SHA-256 over 14 UTF-16BE characters: the 10-character
// blank-padded canonical profile, then the password's rightmost four
// characters (UTF-16 code units), right-padded with spaces.
func level4Salt(profileUTF16, passwordUTF16 []byte) [sha256.Size]byte {
	tail := make([]byte, 0, 8)
	if n := len(passwordUTF16); n >= 8 {
		tail = append(tail, passwordUTF16[n-8:]...)
	} else {
		tail = append(tail, passwordUTF16...)
		for len(tail) < 8 {
			tail = append(tail, 0, ' ')
		}
	}
	h := sha256.New()
	h.Write(profileUTF16)
	h.Write(tail)
	wipe(tail)
	var out [sha256.Size]byte
	h.Sum(out[:0])
	return out
}

// level4Token is PBKDF2-HMAC-SHA512 over the UTF-8 password with the level 4
// salt, 10022 iterations, 64 bytes.
func level4Token(passwordUTF8 []byte, salt [sha256.Size]byte) []byte {
	return pbkdf2.Key(passwordUTF8, salt[:], level4Iterations, sha512.Size, sha512.New)
}

// sha512Substitute is SHA512(token || server seed || client seed || profile
// UTF-16BE || sequence).
func sha512Substitute(token, profileUTF16 []byte, clientSeed, serverSeed [SeedLength]byte) [SHA512SubstituteLength]byte {
	h := sha512.New()
	h.Write(token)
	h.Write(serverSeed[:])
	h.Write(clientSeed[:])
	h.Write(profileUTF16)
	h.Write(sequenceOne[:])
	var out [SHA512SubstituteLength]byte
	h.Sum(out[:0])
	return out
}

// level4Password validates a level 4 password with the same rules as levels
// 2/3 and returns it as UTF-8 and UTF-16BE (both to be wiped by the caller).
func level4Password(password []byte) (utf8Form, utf16Form []byte, err error) {
	utf16Form, err = shaPasswordUTF16BE(password)
	if err != nil {
		return nil, nil, err
	}
	// Re-encode the normalized (trimmed) password as UTF-8.
	units := make([]uint16, len(utf16Form)/2)
	for i := range units {
		units[i] = uint16(utf16Form[2*i])<<8 | uint16(utf16Form[2*i+1])
	}
	runes := utf16.Decode(units)
	wipeUint16(units)
	utf8Form = []byte(string(runes))
	wipeRunes(runes)
	return utf8Form, utf16Form, nil
}
