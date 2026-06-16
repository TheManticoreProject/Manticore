// Package registry holds network-independent helpers for working with the
// Windows registry as data: the REG_* value types, a typed Value carrying a
// REG_* type and its raw little-endian bytes, and (in the regfile sub-package)
// an encoder/decoder for the textual ".reg" export format produced by regedit
// and reg.exe.
//
// Value mirrors the wire representation used by the remote registry protocol
// (MS-RRP): a type tag plus the data bytes exactly as stored. The decode
// helpers (String, Uint32, ...) project the bytes per Type; the constructor
// helpers (StringValue, DwordValue, ...) build the bytes. It is intentionally
// field-compatible with the ms_rrp RegistryValue type so callers can convert
// between the two with a one-line struct literal.
package registry

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"
)

// REG_* registry value types.
//
// [MS-RRP] 2.2.9 / winreg REG_* constants. Values match the numeric type tags
// carried on the wire and written in ".reg" files (e.g. hex(b) is REG_QWORD = 11).
const (
	RegNone                     uint32 = 0  // REG_NONE
	RegSz                       uint32 = 1  // REG_SZ (UTF-16LE, NUL-terminated)
	RegExpandSz                 uint32 = 2  // REG_EXPAND_SZ
	RegBinary                   uint32 = 3  // REG_BINARY
	RegDword                    uint32 = 4  // REG_DWORD (little-endian)
	RegDwordBigEndian           uint32 = 5  // REG_DWORD_BIG_ENDIAN
	RegLink                     uint32 = 6  // REG_LINK
	RegMultiSz                  uint32 = 7  // REG_MULTI_SZ
	RegResourceList             uint32 = 8  // REG_RESOURCE_LIST
	RegFullResourceDescriptor   uint32 = 9  // REG_FULL_RESOURCE_DESCRIPTOR
	RegResourceRequirementsList uint32 = 10 // REG_RESOURCE_REQUIREMENTS_LIST
	RegQword                    uint32 = 11 // REG_QWORD (little-endian)
)

// Value is a typed registry value: its REG_* type and its raw little-endian
// data bytes, exactly as carried on the wire and stored in the registry.
type Value struct {
	Type uint32
	Data []byte
}

// String decodes REG_SZ / REG_EXPAND_SZ / REG_LINK as a UTF-16LE string
// (trailing NULs stripped). For other types it returns the empty string; use
// the type-specific helpers.
func (v Value) String() string {
	switch v.Type {
	case RegSz, RegExpandSz, RegLink:
		return decodeUTF16(v.Data)
	default:
		return ""
	}
}

// Uint32 decodes a REG_DWORD (little-endian). ok is false if the data is too short.
func (v Value) Uint32() (uint32, bool) {
	if len(v.Data) < 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(v.Data), true
}

// Uint64 decodes a REG_QWORD (little-endian). ok is false if the data is too short.
func (v Value) Uint64() (uint64, bool) {
	if len(v.Data) < 8 {
		return 0, false
	}
	return binary.LittleEndian.Uint64(v.Data), true
}

// MultiString decodes a REG_MULTI_SZ: a sequence of UTF-16LE strings, each
// NUL-terminated, the whole terminated by an empty string. Empty trailing
// entries are dropped.
func (v Value) MultiString() []string {
	if v.Type != RegMultiSz {
		return nil
	}
	whole := decodeUTF16Raw(v.Data)
	parts := strings.Split(whole, "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// StringValue builds a REG_SZ value from a Go string.
func StringValue(s string) Value {
	return Value{Type: RegSz, Data: encodeUTF16(s)}
}

// ExpandStringValue builds a REG_EXPAND_SZ value from a Go string.
func ExpandStringValue(s string) Value {
	return Value{Type: RegExpandSz, Data: encodeUTF16(s)}
}

// DwordValue builds a REG_DWORD value (little-endian).
func DwordValue(v uint32) Value {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return Value{Type: RegDword, Data: b}
}

// QwordValue builds a REG_QWORD value (little-endian).
func QwordValue(v uint64) Value {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return Value{Type: RegQword, Data: b}
}

// BinaryValue builds a REG_BINARY value from raw bytes.
func BinaryValue(b []byte) Value {
	return Value{Type: RegBinary, Data: append([]byte(nil), b...)}
}

// NoneValue builds a REG_NONE value from raw bytes.
func NoneValue(b []byte) Value {
	return Value{Type: RegNone, Data: append([]byte(nil), b...)}
}

// MultiStringValue builds a REG_MULTI_SZ value: each string NUL-terminated, the
// block terminated by a final empty string.
func MultiStringValue(items []string) Value {
	return Value{Type: RegMultiSz, Data: encodeUTF16Multi(items)}
}

// --- UTF-16LE helpers ---

// encodeUTF16 encodes s as UTF-16LE with a terminating NUL (the form REG_SZ uses).
func encodeUTF16(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, len(u)*2+2)
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[i*2:], c)
	}
	return b
}

// encodeUTF16Multi encodes a REG_MULTI_SZ block: each item NUL-terminated, a
// final empty string closing the block.
func encodeUTF16Multi(items []string) []byte {
	var b []byte
	for _, s := range items {
		b = append(b, encodeUTF16(s)...)
	}
	b = append(b, 0x00, 0x00) // final empty string terminator
	return b
}

// decodeUTF16 decodes UTF-16LE bytes and strips a single trailing NUL terminator.
func decodeUTF16(b []byte) string {
	return strings.TrimRight(decodeUTF16Raw(b), "\x00")
}

// decodeUTF16Raw decodes UTF-16LE bytes without stripping NULs (an odd trailing
// byte is ignored).
func decodeUTF16Raw(b []byte) string {
	n := len(b) / 2
	u := make([]uint16, n)
	for i := 0; i < n; i++ {
		u[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return string(utf16.Decode(u))
}
