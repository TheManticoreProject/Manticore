package ebcdic

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

// ErrUnmappable is matched (via errors.Is) by every Codec encoding failure.
var ErrUnmappable = errors.New("ebcdic: text not representable in code page")

// UnmappableError reports the first input rune a Codec cannot encode. Invalid
// UTF-8 is reported with Rune set to utf8.RuneError.
type UnmappableError struct {
	Codepage string
	Rune     rune
	Offset   int // byte offset in the input string
}

func (e *UnmappableError) Error() string {
	if e.Rune == utf8.RuneError {
		return fmt.Sprintf("ebcdic: invalid UTF-8 at byte %d for %s", e.Offset, e.Codepage)
	}
	return fmt.Sprintf("ebcdic: %U at byte %d is not representable in %s", e.Rune, e.Offset, e.Codepage)
}

// Is makes errors.Is(err, ErrUnmappable) true.
func (e *UnmappableError) Is(target error) bool { return target == ErrUnmappable }

// Codec converts between Unicode strings and one single-byte EBCDIC code
// page. Unlike EbcdicToAscii and AsciiToEbcdic it never modifies its input,
// always returns freshly allocated output, and rejects text it cannot
// represent instead of substituting a replacement byte.
//
// A Codec is immutable and safe for concurrent use.
type Codec struct {
	name string
	cm   *charmap.Charmap
}

var (
	codecCP037  = &Codec{name: "cp037", cm: charmap.CodePage037}
	codecCP1140 = &Codec{name: "cp1140", cm: charmap.CodePage1140}
)

// CP037 returns the codec for IBM code page 037 (US/Canada EBCDIC).
func CP037() *Codec { return codecCP037 }

// CP1140 returns the codec for IBM code page 1140: code page 037 with the
// euro sign at 0x9F in place of the currency sign.
func CP1140() *Codec { return codecCP1140 }

// Name returns the code page name, for example "cp037".
func (c *Codec) Name() string { return c.name }

// Encode converts s to EBCDIC. It fails on invalid UTF-8 and on any rune the
// code page cannot represent; it never substitutes or drops characters.
func (c *Codec) Encode(s string) ([]byte, error) {
	out := make([]byte, 0, len(s))
	for i, r := range s {
		if r == utf8.RuneError {
			if _, size := utf8.DecodeRuneInString(s[i:]); size == 1 {
				return nil, &UnmappableError{Codepage: c.name, Rune: utf8.RuneError, Offset: i}
			}
		}
		b, ok := c.cm.EncodeRune(r)
		if !ok {
			return nil, &UnmappableError{Codepage: c.name, Rune: r, Offset: i}
		}
		out = append(out, b)
	}
	return out, nil
}

// Decode converts EBCDIC bytes to a string. Every byte of these code pages
// maps to exactly one rune, so decoding cannot fail; the error result keeps
// the signature uniform with Encode.
func (c *Codec) Decode(b []byte) (string, error) {
	var sb strings.Builder
	sb.Grow(len(b))
	for _, v := range b {
		sb.WriteRune(c.cm.DecodeByte(v))
	}
	return sb.String(), nil
}
