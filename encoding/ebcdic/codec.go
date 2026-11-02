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
	// Exactly one of cm and table is set: x/text carries 037, 1047 and 1140,
	// and the other IBM code pages come from the generated tables.
	cm     *charmap.Charmap
	table  *[256]rune
	encode map[rune]byte
}

var (
	codecCP037  = &Codec{name: "cp037", cm: charmap.CodePage037}
	codecCP1140 = &Codec{name: "cp1140", cm: charmap.CodePage1140}
)

// ibmCodecs holds the codec of every IBM code page IBMCodePage knows, built once.
var ibmCodecs = func() map[int]*Codec {
	codecs := map[int]*Codec{
		37:   codecCP037,
		1047: {name: "cp1047", cm: charmap.CodePage1047},
		1140: codecCP1140,
	}
	for cp, table := range ibmTables {
		encode := make(map[rune]byte, len(table))
		for b, r := range table {
			encode[r] = byte(b)
		}
		codecs[cp] = &Codec{name: fmt.Sprintf("cp%03d", cp), table: table, encode: encode}
	}
	return codecs
}()

// IBMCodePage returns the codec for the single-byte EBCDIC IBM code page cp,
// such as 273 (Germany) or 297 (France), and false when the package carries no
// table for it. Code pages 037, 1047 and 1140 use golang.org/x/text; the others
// are generated from IBM's mappings as published by ICU, and both directions
// use only round-trip mappings.
func IBMCodePage(cp int) (*Codec, bool) {
	c, ok := ibmCodecs[cp]
	return c, ok
}

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
		b, ok := c.encodeRune(r)
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
		if c.table != nil {
			sb.WriteRune(c.table[v])
		} else {
			sb.WriteRune(c.cm.DecodeByte(v))
		}
	}
	return sb.String(), nil
}

// encodeRune maps one rune to its byte in the code page.
func (c *Codec) encodeRune(r rune) (byte, bool) {
	if c.table != nil {
		b, ok := c.encode[r]
		return b, ok
	}
	return c.cm.EncodeRune(r)
}
