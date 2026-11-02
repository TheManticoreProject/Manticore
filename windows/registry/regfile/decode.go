package regfile

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"

	"github.com/TheManticoreProject/Manticore/windows/registry"
)

// ValueLine is one value entry within a key block. Delete is true for a delete
// directive (`"Name"=-`), in which case Value is the zero Value. Name is empty
// for the default value (the "@" token).
type ValueLine struct {
	Name   string
	Value  registry.Value
	Delete bool
}

// KeyBlock is a key and its values, as written between two key headers in a
// ".reg" file. Delete is true for a key-delete directive (`[-HKEY_...\Key]`),
// in which case Values is empty.
type KeyBlock struct {
	Path   string
	Delete bool
	Values []ValueLine
}

// Parse decodes a ".reg" file into its key blocks. It is tolerant of real-world
// inputs: it auto-detects UTF-16LE/UTF-16BE (with BOM) and UTF-8 (with or
// without BOM) encodings, accepts both the "Windows Registry Editor Version
// 5.00" and legacy "REGEDIT4" headers, ignores comment (";") and blank lines,
// folds backslash line-continuations in long hex values, and understands the
// delete directives "[-Key]" and `"Name"=-`.
//
// Input without a BOM that is not valid UTF-8 is a legacy ANSI file, as regedit
// writes REGEDIT4 exports ("regedit /a"), and is decoded from Windows-1252; use
// ParseWithCodePage for another code page. Input that is valid UTF-8 is read as
// UTF-8, so an ANSI file whose bytes happen to form valid UTF-8 sequences (such
// as C3 A9) is taken as UTF-8.
//
// Parsing is best-effort: lines that cannot be interpreted are skipped and the
// blocks parsed so far are still returned, with the per-line problems joined
// into the returned error. A nil error means every non-trivial line parsed.
func Parse(raw []byte) ([]KeyBlock, error) {
	return ParseWithCodePage(raw, charmap.Windows1252)
}

// ParseWithCodePage is Parse for legacy ANSI ".reg" files written in the given
// Windows code page (e.g. charmap.CodePage850, charmap.Windows1251).
//
// The code page applies in two places. Text without a BOM that is not valid
// UTF-8 is decoded from it. And in a REGEDIT4 file, regedit writes the data of
// string-typed hex values (hex(1) REG_SZ, hex(2) REG_EXPAND_SZ, hex(7)
// REG_MULTI_SZ) as code-page bytes rather than UTF-16LE, so that data is
// converted to the UTF-16LE form those registry types hold. A nil codePage
// selects Windows-1252.
func ParseWithCodePage(raw []byte, codePage encoding.Encoding) ([]KeyBlock, error) {
	if codePage == nil {
		codePage = charmap.Windows1252
	}
	text := decodeText(raw, codePage)
	lines := foldContinuations(splitLines(text))

	var blocks []KeyBlock
	var errs []error
	cur := -1 // index into blocks of the key currently receiving values
	// ansi is set once a REGEDIT4 header is seen: its string-typed hex data is
	// in the code page.
	var ansi encoding.Encoding

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			continue
		case strings.HasPrefix(trimmed, ";"):
			continue
		case trimmed == "REGEDIT4":
			ansi = codePage
			continue
		case strings.HasPrefix(trimmed, "Windows Registry Editor Version"):
			// A version 5 section (e.g. after concatenated files) holds
			// UTF-16LE hex data again.
			ansi = nil
			continue
		case strings.HasPrefix(trimmed, "["):
			path, del, ok := parseKeyHeader(trimmed)
			if !ok {
				errs = append(errs, fmt.Errorf("line %d: malformed key header %q", i+1, trimmed))
				continue
			}
			blocks = append(blocks, KeyBlock{Path: path, Delete: del})
			cur = len(blocks) - 1
		default:
			if cur < 0 {
				errs = append(errs, fmt.Errorf("line %d: value line before any key: %q", i+1, trimmed))
				continue
			}
			vl, fromHex, err := parseValueLine(trimmed)
			if err == nil && ansi != nil && fromHex {
				vl.Value, err = ansiStringDataToUTF16(vl.Value, ansi)
			}
			if err != nil {
				errs = append(errs, fmt.Errorf("line %d: %w", i+1, err))
				continue
			}
			blocks[cur].Values = append(blocks[cur].Values, vl)
		}
	}

	return blocks, errors.Join(errs...)
}

// parseKeyHeader parses a "[...]" key line, returning the inner path and whether
// it is a delete directive ("[-...]").
func parseKeyHeader(line string) (path string, del bool, ok bool) {
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return "", false, false
	}
	inner := line[1 : len(line)-1]
	if strings.HasPrefix(inner, "-") {
		return inner[1:], true, true
	}
	return inner, false, true
}

// parseValueLine parses a "name"=data (or @=data) line into a ValueLine.
// fromHex reports that the data was written in the hex(N): form, whose bytes
// are taken verbatim (and so, in a REGEDIT4 file, are code-page bytes for a
// string type).
func parseValueLine(line string) (vl ValueLine, fromHex bool, err error) {
	name, rest, err := splitNameValue(line)
	if err != nil {
		return ValueLine{}, false, err
	}
	v, del, err := parseValueData(rest)
	if err != nil {
		return ValueLine{}, false, err
	}
	return ValueLine{Name: name, Value: v, Delete: del}, hasPrefixFold(rest, "hex("), nil
}

// splitNameValue splits a value line into its name and the data text after the
// "=". The default value is written as "@"; named values are quoted and may
// contain escaped quotes/backslashes.
func splitNameValue(line string) (name, rest string, err error) {
	switch {
	case strings.HasPrefix(line, "@"):
		after := strings.TrimSpace(line[1:])
		if !strings.HasPrefix(after, "=") {
			return "", "", fmt.Errorf("expected '=' after '@'")
		}
		return "", strings.TrimSpace(after[1:]), nil
	case strings.HasPrefix(line, `"`):
		n, after, ok := scanQuoted(line)
		if !ok {
			return "", "", fmt.Errorf("unterminated value name")
		}
		after = strings.TrimSpace(after)
		if !strings.HasPrefix(after, "=") {
			return "", "", fmt.Errorf("expected '=' after value name")
		}
		return n, strings.TrimSpace(after[1:]), nil
	default:
		return "", "", fmt.Errorf("malformed value line: %q", line)
	}
}

// ansiStringDataToUTF16 converts the data of a REGEDIT4 string-typed hex(N)
// value (REG_SZ, REG_EXPAND_SZ, REG_MULTI_SZ) from code-page bytes to UTF-16LE.
// It is only applied to the hex(N): form: a quoted "..." string is decoded text
// and already UTF-16LE in the Value. Other types are binary and left alone.
func ansiStringDataToUTF16(v registry.Value, codePage encoding.Encoding) (registry.Value, error) {
	switch v.Type {
	case registry.RegSz, registry.RegExpandSz, registry.RegMultiSz:
	default:
		return v, nil
	}
	text, err := codePage.NewDecoder().Bytes(v.Data)
	if err != nil {
		return v, fmt.Errorf("decode hex(%x) string data: %w", v.Type, err)
	}
	units := utf16.Encode([]rune(string(text)))
	data := make([]byte, 2*len(units))
	for i, u := range units {
		binary.LittleEndian.PutUint16(data[2*i:], u)
	}
	v.Data = data
	return v, nil
}

// hasPrefixFold reports whether s begins with prefix, ignoring ASCII case:
// regedit writes "hex(2):" but hand-edited files use "HEX(2):" or "Hex:".
func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

// parseValueData parses the right-hand side of a value line into a typed value.
// del is true for the delete directive ("-").
func parseValueData(rest string) (v registry.Value, del bool, err error) {
	switch {
	case rest == "-":
		return registry.Value{}, true, nil
	case strings.HasPrefix(rest, `"`):
		s, _, ok := scanQuoted(rest)
		if !ok {
			return registry.Value{}, false, fmt.Errorf("unterminated string value")
		}
		return registry.StringValue(s), false, nil
	case hasPrefixFold(rest, "dword:"):
		n, err := strconv.ParseUint(strings.TrimSpace(rest[len("dword:"):]), 16, 32)
		if err != nil {
			return registry.Value{}, false, fmt.Errorf("invalid dword: %w", err)
		}
		return registry.DwordValue(uint32(n)), false, nil
	case hasPrefixFold(rest, "hex("):
		closeIdx := strings.Index(rest, ")")
		if closeIdx < 0 {
			return registry.Value{}, false, fmt.Errorf("malformed hex(N) value: %q", rest)
		}
		t, err := strconv.ParseUint(strings.TrimSpace(rest[len("hex("):closeIdx]), 16, 32)
		if err != nil {
			return registry.Value{}, false, fmt.Errorf("invalid hex type: %w", err)
		}
		after := strings.TrimPrefix(strings.TrimSpace(rest[closeIdx+1:]), ":")
		data, err := parseHexBytes(after)
		if err != nil {
			return registry.Value{}, false, err
		}
		return registry.Value{Type: uint32(t), Data: data}, false, nil
	case hasPrefixFold(rest, "hex:"):
		data, err := parseHexBytes(rest[len("hex:"):])
		if err != nil {
			return registry.Value{}, false, err
		}
		return registry.Value{Type: registry.RegBinary, Data: data}, false, nil
	default:
		return registry.Value{}, false, fmt.Errorf("unrecognized value data: %q", rest)
	}
}

// parseHexBytes parses a comma-separated list of two-digit hex bytes. Whitespace
// and empty fields (from trailing commas) are ignored. An empty list yields an
// empty, non-nil slice.
func parseHexBytes(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	// Keep nil for the empty list so values round-trip against the registry
	// constructor helpers (BinaryValue([]byte{}) etc.), which also yield nil.
	var out []byte
	if s == "" {
		return out, nil
	}
	for _, tok := range strings.Split(s, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		b, err := strconv.ParseUint(tok, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex byte %q: %w", tok, err)
		}
		out = append(out, byte(b))
	}
	return out, nil
}

// scanQuoted reads a double-quoted token at the start of s, unescaping `\\` and
// `\"`. It returns the unescaped contents, the text following the closing quote,
// and ok=false if s does not start with a quote or is unterminated.
func scanQuoted(s string) (inner, rest string, ok bool) {
	if len(s) == 0 || s[0] != '"' {
		return "", s, false
	}
	var b strings.Builder
	for i := 1; i < len(s); {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s):
			n := s[i+1]
			if n == '\\' || n == '"' {
				b.WriteByte(n)
				i += 2
				continue
			}
			// Lone backslash (not a recognized escape): keep it literally.
			b.WriteByte('\\')
			i++
		case c == '"':
			return b.String(), s[i+1:], true
		default:
			b.WriteByte(c)
			i++
		}
	}
	return "", s, false
}

// splitLines normalizes CRLF/CR/LF endings and splits the text into lines.
func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}

// foldContinuations joins lines split with a trailing backslash (regedit's long
// hex-value wrapping) into single logical lines. Leading whitespace on each
// continuation segment is stripped so the rejoined byte list parses cleanly.
//
// A backslash only continues a line within a (possibly multi-line) hex value,
// so a continuation is never *started* from a comment, header, or key line: a
// trailing backslash there (e.g. a path at the end of a comment) is literal.
// Once a continuation is in progress, every following physical line is folded
// until one does not end with a backslash.
func foldContinuations(lines []string) []string {
	var out []string
	var cur strings.Builder
	pending := false

	for _, ln := range lines {
		body := strings.TrimRight(ln, " \t")
		continues := strings.HasSuffix(body, "\\")
		if continues && !pending && !canStartContinuation(strings.TrimSpace(ln)) {
			continues = false
		}
		if continues {
			body = strings.TrimSuffix(body, "\\")
		}
		if pending {
			cur.WriteString(strings.TrimLeft(body, " \t"))
		} else {
			cur.WriteString(body)
		}
		if continues {
			pending = true
			continue
		}
		out = append(out, cur.String())
		cur.Reset()
		pending = false
	}
	if pending { // dangling continuation with no terminating line
		out = append(out, cur.String())
	}
	return out
}

// canStartContinuation reports whether a logical line may begin a backslash
// continuation. Only value lines wrap in regedit output; comment, header, and
// key lines never do, so a trailing backslash on those is treated as literal.
func canStartContinuation(trimmed string) bool {
	switch {
	case strings.HasPrefix(trimmed, ";"):
		return false
	case strings.HasPrefix(trimmed, "["):
		return false
	case trimmed == "REGEDIT4":
		return false
	case strings.HasPrefix(trimmed, "Windows Registry Editor Version"):
		return false
	default:
		return true
	}
}

// decodeText sniffs the byte-order mark / encoding of a ".reg" file and returns
// its text as a Go (UTF-8) string. It recognizes UTF-16LE and UTF-16BE (with
// BOM) and UTF-8 (with or without BOM). Input without a BOM that is not valid
// UTF-8 is a legacy ANSI file and is decoded from codePage.
func decodeText(raw []byte, codePage encoding.Encoding) string {
	if codePage == nil {
		codePage = charmap.Windows1252
	}
	switch {
	case len(raw) >= 2 && raw[0] == 0xFF && raw[1] == 0xFE:
		return decodeUTF16(raw[2:], binary.LittleEndian)
	case len(raw) >= 2 && raw[0] == 0xFE && raw[1] == 0xFF:
		return decodeUTF16(raw[2:], binary.BigEndian)
	case len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF:
		return string(raw[3:])
	case utf8.Valid(raw):
		return string(raw)
	default:
		text, err := codePage.NewDecoder().Bytes(raw)
		if err != nil {
			return string(raw)
		}
		return string(text)
	}
}

// decodeUTF16 decodes UTF-16 bytes in the given byte order to a Go string (an
// odd trailing byte is ignored).
func decodeUTF16(b []byte, order binary.ByteOrder) string {
	n := len(b) / 2
	u := make([]uint16, n)
	for i := 0; i < n; i++ {
		u[i] = order.Uint16(b[i*2:])
	}
	return string(utf16.Decode(u))
}
