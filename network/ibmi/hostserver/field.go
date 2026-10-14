package hostserver

import (
	"encoding/binary"
	"fmt"
	"math"
)

// FieldPrefixLength is the size of an LL/CP field's length and code point.
const FieldPrefixLength = 6

// DefaultMaxFields is the field-count limit used when FieldPolicy.MaxFields
// is 0.
const DefaultMaxFields = 256

// Field is one LL/CP variable field: a 32-bit big-endian length that
// includes the 6-byte prefix, a 16-bit code point, then the value.
type Field struct {
	CodePoint uint16
	Value     []byte
}

// AppendField appends the encoding of one field to dst.
func AppendField(dst []byte, codePoint uint16, value []byte) ([]byte, error) {
	total := uint64(FieldPrefixLength) + uint64(len(value))
	if total > math.MaxUint32 {
		return nil, fmt.Errorf("%w: field of %d bytes", ErrFrameTooLarge, total)
	}
	var prefix [FieldPrefixLength]byte
	binary.BigEndian.PutUint32(prefix[0:], uint32(total))
	binary.BigEndian.PutUint16(prefix[4:], codePoint)
	dst = append(dst, prefix[:]...)
	return append(dst, value...), nil
}

// FieldPolicy bounds LL/CP parsing.
type FieldPolicy struct {
	// MaxFields caps the number of fields; 0 means DefaultMaxFields.
	MaxFields int
	// Repeatable lists code points that may occur more than once. Every
	// other code point is a singleton and a duplicate is rejected.
	Repeatable map[uint16]bool
}

// Fields is an ordered list of parsed fields.
type Fields []Field

// Get returns the value of the first field with codePoint.
func (fs Fields) Get(codePoint uint16) ([]byte, bool) {
	for _, f := range fs {
		if f.CodePoint == codePoint {
			return f.Value, true
		}
	}
	return nil, false
}

// All returns the values of every field with codePoint, in order.
func (fs Fields) All(codePoint uint16) [][]byte {
	var out [][]byte
	for _, f := range fs {
		if f.CodePoint == codePoint {
			out = append(out, f.Value)
		}
	}
	return out
}

// ParseFields decodes b as a sequence of LL/CP fields that must consume it
// exactly. Values are sub-slices of b. Errors report the byte offset and code
// point, never the field contents.
func ParseFields(b []byte, policy FieldPolicy) (Fields, error) {
	max := policy.MaxFields
	if max <= 0 {
		max = DefaultMaxFields
	}
	var out Fields
	seen := map[uint16]bool{}
	for off := 0; off < len(b); {
		if len(b)-off < FieldPrefixLength {
			return nil, malformed("field", off, "%d trailing bytes cannot hold a field prefix", len(b)-off)
		}
		ll := binary.BigEndian.Uint32(b[off:])
		cp := binary.BigEndian.Uint16(b[off+4:])
		if ll < FieldPrefixLength {
			return nil, malformed("field", off, "length %d of code point 0x%04X is below %d", ll, cp, FieldPrefixLength)
		}
		if uint64(ll) > uint64(len(b)-off) {
			return nil, malformed("field", off, "length %d of code point 0x%04X exceeds the %d remaining bytes", ll, cp, len(b)-off)
		}
		if len(out) == max {
			return nil, malformed("field", off, "more than %d fields", max)
		}
		if seen[cp] && !policy.Repeatable[cp] {
			return nil, malformed("field", off, "duplicate singleton code point 0x%04X", cp)
		}
		seen[cp] = true
		end := off + int(ll)
		out = append(out, Field{CodePoint: cp, Value: b[off+FieldPrefixLength : end : end]})
		off = end
	}
	return out, nil
}
