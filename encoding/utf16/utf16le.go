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
//
// A trailing odd byte is an incomplete code unit and is dropped rather than
// decoded. Input of an odd length is not a caller's mistake to be punished with a
// panic: this decoder is fed by fields taken off the network, where a length is
// whatever the sender said it was, and a truncated final character is the ordinary
// consequence of a field that was cut short.
func DecodeUTF16LE(b []byte) string {
	utf16le := make([]uint16, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		utf16le[i/2] = uint16(b[i]) | (uint16(b[i+1]) << 8)
	}
	return string(utf16.Decode(utf16le))
}

// IsUTF16LE checks if a byte slice is valid UTF-16LE encoded data
//
// This function takes a byte slice and checks if it represents valid UTF-16LE encoded data.
// It verifies that:
// 1. The length is even (UTF-16 uses 2 bytes per code unit)
// 2. Each code unit has 0x00 in the high byte, indicating ASCII range characters
//
// Note: This is a simplified check that only validates ASCII range characters.
// For full UTF-16LE validation, surrogate pairs and non-ASCII characters would need to be handled.
//
// Returns:
//   - true if the byte slice appears to be valid UTF-16LE ASCII text
//   - false otherwise
func IsUTF16LE(b []byte) bool {
	// Check length is even
	if len(b)%2 != 0 {
		return false
	}

	// Check that every other byte is 0x00
	for i := 1; i < len(b); i += 2 {
		if b[i] != 0x00 {
			return false
		}
	}

	return true
}
