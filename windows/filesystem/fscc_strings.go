package filesystem

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"
)

// encodeUTF16LE encodes s as little-endian UTF-16, the on-wire form of the
// variable name fields in MS-FSCC information-class structures.
func encodeUTF16LE(s string) []byte {
	u16 := utf16.Encode([]rune(s))
	b := make([]byte, len(u16)*2)
	for i, u := range u16 {
		binary.LittleEndian.PutUint16(b[i*2:], u)
	}
	return b
}

// decodeUTF16LE decodes a little-endian UTF-16 byte buffer into a string,
// trimming any trailing NUL units.
func decodeUTF16LE(b []byte) string {
	u16 := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u16 = append(u16, binary.LittleEndian.Uint16(b[i:i+2]))
	}
	return strings.TrimRight(string(utf16.Decode(u16)), "\x00")
}
