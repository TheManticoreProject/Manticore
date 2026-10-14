package ccsid

import (
	"errors"
	"fmt"
)

// Blank is the EBCDIC space used to pad fixed-width IBM i character fields.
const Blank byte = 0x40

// ErrFieldTooLong is matched (via errors.Is) when text does not fit a
// fixed-width field. Text is never truncated.
var ErrFieldTooLong = errors.New("ccsid: text longer than fixed-width field")

// EncodeFixed encodes s with c and pads it on the right with EBCDIC blanks
// to exactly width bytes. It fails, without truncating, if the encoded text
// is longer than width or contains an unrepresentable rune.
func EncodeFixed(c Codec, s string, width int) ([]byte, error) {
	if width < 0 {
		return nil, fmt.Errorf("ccsid: negative field width %d", width)
	}
	enc, err := c.Encode(s)
	if err != nil {
		return nil, err
	}
	if len(enc) > width {
		return nil, fmt.Errorf("%w: %d bytes for a %d-byte field", ErrFieldTooLong, len(enc), width)
	}
	out := make([]byte, width)
	copy(out, enc)
	for i := len(enc); i < width; i++ {
		out[i] = Blank
	}
	return out, nil
}

// DecodeFixed decodes a fixed-width field with c after removing trailing
// EBCDIC blanks (0x40). Leading blanks and any other trailing byte, including
// 0x00, are data and are kept. The input is never modified.
func DecodeFixed(c Codec, field []byte) (string, error) {
	end := len(field)
	for end > 0 && field[end-1] == Blank {
		end--
	}
	return c.Decode(field[:end])
}
