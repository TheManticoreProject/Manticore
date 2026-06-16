// Package regfile encodes and decodes the textual ".reg" registry export
// format produced by regedit and reg.exe (the "Windows Registry Editor Version
// 5.00" / legacy "REGEDIT4" files).
//
// It is pure, network-independent logic built on the registry.Value type: the
// encoder turns key/value blocks into the textual format, and the (tolerant)
// decoder parses real-world ".reg" files back into structured blocks. Together
// they let callers implement reg.exe-style EXPORT/IMPORT on top of any registry
// backend.
package regfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/windows/registry"
)

// Header is the first line of a version-5 (".reg" v5.00) file. It is written in
// UTF-16LE with a byte-order mark by Marshal.
const Header = "Windows Registry Editor Version 5.00"

// maxHexLineLen is the column past which the encoder wraps a long hex byte list
// onto a continuation line (ending the prior line with a backslash), matching
// regedit's ~80-column wrapping. It is cosmetic: the decoder folds continuations
// regardless of width.
const maxHexLineLen = 80

// escaper escapes the two characters that are special inside a quoted token
// (value name or REG_SZ data): backslash and double-quote. Backslash must be
// replaced first.
var escaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

// FormatKeyHeader returns the bracketed key line for path, e.g.
// "[HKEY_LOCAL_MACHINE\Software\Foo]". When del is true it emits the delete
// directive "[-HKEY_LOCAL_MACHINE\Software\Foo]". Backslashes in the path are
// not escaped (they are literal path separators).
func FormatKeyHeader(path string, del bool) string {
	if del {
		return "[-" + path + "]"
	}
	return "[" + path + "]"
}

// FormatValueLine returns a single ".reg" value line for the named value, e.g.
// `"Name"="data"` or `@=dword:0000001f` for the default value (name == "").
// The data token is chosen from v.Type per the ".reg" type encodings.
func FormatValueLine(name string, v registry.Value) string {
	return nameToken(name) + "=" + valueToken(v)
}

// FormatDeleteValueLine returns the directive that deletes the named value:
// `"Name"=-` (or `@=-` for the default value).
func FormatDeleteValueLine(name string) string {
	return nameToken(name) + "=-"
}

// Marshal renders the blocks as a complete ".reg" file: the version-5 header, a
// blank line, then each key block (its header followed by one line per value).
// The result is UTF-16LE with a leading byte-order mark and CRLF line endings,
// exactly as regedit writes exports.
func Marshal(blocks []KeyBlock) []byte {
	var b strings.Builder
	b.WriteString(Header)
	b.WriteString("\r\n\r\n")
	for _, blk := range blocks {
		b.WriteString(FormatKeyHeader(blk.Path, blk.Delete))
		b.WriteString("\r\n")
		if !blk.Delete {
			for _, vl := range blk.Values {
				if vl.Delete {
					b.WriteString(FormatDeleteValueLine(vl.Name))
				} else {
					b.WriteString(FormatValueLine(vl.Name, vl.Value))
				}
				b.WriteString("\r\n")
			}
		}
		b.WriteString("\r\n")
	}
	return encodeUTF16LEWithBOM(b.String())
}

// nameToken renders the left-hand side of a value line: "@" for the default
// value, otherwise the escaped, quoted name.
func nameToken(name string) string {
	if name == "" {
		return "@"
	}
	return `"` + escaper.Replace(name) + `"`
}

// valueToken renders the right-hand side of a value line per v.Type.
func valueToken(v registry.Value) string {
	switch v.Type {
	case registry.RegSz:
		return `"` + escaper.Replace(v.String()) + `"`
	case registry.RegDword:
		if n, ok := v.Uint32(); ok {
			return fmt.Sprintf("dword:%08x", n)
		}
		return formatHexBytes("hex(4):", v.Data)
	case registry.RegBinary:
		return formatHexBytes("hex:", v.Data)
	case registry.RegQword:
		return formatHexBytes("hex(b):", v.Data)
	case registry.RegExpandSz:
		return formatHexBytes("hex(2):", v.Data)
	case registry.RegMultiSz:
		return formatHexBytes("hex(7):", v.Data)
	case registry.RegNone:
		return formatHexBytes("hex(0):", v.Data)
	default:
		return formatHexBytes(fmt.Sprintf("hex(%x):", v.Type), v.Data)
	}
}

// formatHexBytes renders data as a comma-separated lowercase hex byte list
// prefixed by prefix (e.g. "hex:" or "hex(7):"), wrapping long lists onto
// continuation lines. Each wrapped line ends with a backslash and the next line
// is indented two spaces, matching regedit. An empty data slice yields just the
// prefix (e.g. "hex(0):").
func formatHexBytes(prefix string, data []byte) string {
	if len(data) == 0 {
		return prefix
	}
	var b strings.Builder
	line := prefix
	for i, by := range data {
		tok := fmt.Sprintf("%02x", by)
		if i < len(data)-1 {
			tok += ","
		}
		// Reserve one column for the trailing continuation backslash.
		if len(line)+len(tok)+1 > maxHexLineLen {
			b.WriteString(line)
			b.WriteString("\\\r\n")
			line = "  " + tok
		} else {
			line += tok
		}
	}
	b.WriteString(line)
	return b.String()
}

// encodeUTF16LEWithBOM encodes s as UTF-16LE prefixed by the little-endian BOM
// (0xFFFE), the on-disk form of a version-5 ".reg" file.
func encodeUTF16LEWithBOM(s string) []byte {
	u := utf16.Encode([]rune(s))
	var buf bytes.Buffer
	buf.Grow(2 + len(u)*2)
	buf.Write([]byte{0xFF, 0xFE}) // UTF-16LE BOM
	tmp := make([]byte, 2)
	for _, c := range u {
		binary.LittleEndian.PutUint16(tmp, c)
		buf.Write(tmp)
	}
	return buf.Bytes()
}
