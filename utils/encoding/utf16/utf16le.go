package utf16

import (
	"unicode/utf16"
)

// EncodeUTF16LE encodes a string to UTF-16 little endian bytes
//
// This function takes a string and encodes it to UTF-16 little endian bytes.
// It uses the utf16.Encode function to convert the string to a slice of uint16 values.
// Then, it creates a byte slice with twice the length of the utf16le slice,
// and fills it with the UTF-16 little endian encoded values.
//
// Returns:
//   - A byte slice representing the UTF-16 little endian encoded string.
func EncodeUTF16LE(s string) []byte {
	utf16le := utf16.Encode([]rune(s))
	bytes := make([]byte, len(utf16le)*2)
	for i, r := range utf16le {
		bytes[i*2] = byte(r)
		bytes[i*2+1] = byte(r >> 8)
	}
	return bytes
}

// DecodeUTF16LE decodes UTF-16 little endian bytes to a string
//
// This function takes a byte slice and decodes it to a string.
// It creates a slice of uint16 values with half the length of the byte slice,
// and fills it with the UTF-16 little endian decoded values.
// Then, it uses the utf16.Decode function to convert the uint16 slice to a string.
//
// Returns:
//   - A string representing the UTF-16 little endian decoded byte slice.
func DecodeUTF16LE(b []byte) string {
	utf16le := make([]uint16, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		utf16le[i/2] = uint16(b[i]) | (uint16(b[i+1]) << 8)
	}
	return string(utf16.Decode(utf16le))
}
