package dtyp

import "unicode/utf16"

// RPC_UNICODE_STRING is the [MS-DTYP] 2.3.10 counted Unicode string. It is the single
// most common string carrier in the LSA/SAMR interfaces and, unlike a bare NDR wide
// string, cannot be modeled by ndr.WSTR because of a byte-vs-char gotcha:
//
//   - Length and MaximumLength are counts in BYTES (each must be even), whereas
//   - Buffer is a [unique] pointer to a conformant-varying array of wchar whose
//     maximum_count is MaximumLength/2 and actual_count is Length/2 — counts in CHARS.
//
// The buffer carries no NUL terminator (it is a [size_is/length_is] array, not an NDR
// [string]). Modeling Buffer as a []uint16 tagged "unique,varying" lets the walker emit
// the referent id inline and the char-counted conformant-varying body deferred, so the
// buffers of an array of RPC_UNICODE_STRING are correctly emitted after the whole array
// (see issue #419). Use NewUnicodeString to set the byte counts and buffer together.
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/94a16bb6-c610-4cb9-8db6-26f15f560061
type RPC_UNICODE_STRING struct {
	Length        uint16 // length of the string in bytes, excluding any terminator
	MaximumLength uint16 // size of Buffer in bytes
	Buffer        []uint16 `ndr:"unique,varying"`
}

// NewUnicodeString builds an RPC_UNICODE_STRING from a Go string, encoding it to UTF-16
// and setting Length and MaximumLength to the byte count (2 per code unit). An empty
// string yields a zero-length value with a NULL Buffer, matching the Windows
// representation of an empty counted string.
func NewUnicodeString(s string) RPC_UNICODE_STRING {
	if s == "" {
		return RPC_UNICODE_STRING{}
	}
	units := utf16.Encode([]rune(s))
	n := uint16(len(units) * 2)
	return RPC_UNICODE_STRING{Length: n, MaximumLength: n, Buffer: units}
}

// String decodes the Buffer (truncated to Length/2 code units, as Length governs the
// valid portion) back to a Go string.
func (u RPC_UNICODE_STRING) String() string {
	units := u.Buffer
	if n := int(u.Length / 2); n < len(units) {
		units = units[:n]
	}
	return string(utf16.Decode(units))
}
