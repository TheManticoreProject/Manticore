// Package ndr implements the DCE/RPC Network Data Representation (NDR) transfer
// syntax, version 2.0 (the "NDR20" little-endian encoding used by Windows RPC), with
// a declarative, reflection-driven API for marshalling RPC call structures.
//
// The low-level codec (Encoder/Decoder) handles NDR's stream-relative alignment and
// primitive encoding; the reflection walker in marshal.go drives it from struct tags
// so callers can declare an RPC call as a Go struct (see Call, Marshal, Unmarshal).
//
// The codec is transfer-syntax aware: an Encoder/Decoder carries a Syntax (NDR20 or
// NDR64) that governs the width and alignment of NDR counts and referent ids (4 octets
// 4-aligned for NDR20, 8 octets 8-aligned for NDR64; [MS-RPCE] section 2.2.5). Scalar
// primitives keep their natural width in both syntaxes. The declarative walker
// (Marshal/Unmarshal) and the v5 client still default to NDR20; the NDR64 path is being
// introduced incrementally and is reachable through NewEncoderForSyntax/
// NewDecoderForSyntax. Both syntaxes are little-endian (the standard Windows data
// representation, DREP 0x10); big-endian is out of scope.
//
// References:
//   - [C706] DCE 1.1: RPC, Chapter 14 "Transfer Syntax NDR":
//     https://pubs.opengroup.org/onlinepubs/9629399/chap14.htm
//   - [MS-RPCE] Transfer Syntax NDR types:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/b6090c2b-f44a-47a1-a13b-b82ade0137b2
package ndr

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// Marshaler is implemented by types that encode themselves as NDR, bypassing the
// reflection walker. It is the escape hatch for unions and any layout the declarative
// tags do not cover.
type Marshaler interface {
	// AlignmentNDR reports the type's NDR alignment (1, 2, 4, or 8).
	AlignmentNDR() int
	// MarshalNDR appends the type's NDR representation to the encoder.
	MarshalNDR(*Encoder) error
	// UnmarshalNDR reads the type's NDR representation from the decoder.
	UnmarshalNDR(*Decoder) error
}

// Syntax selects the NDR transfer syntax an Encoder/Decoder operates under. It governs
// the on-the-wire width and alignment of NDR counts (maximum_count, offset,
// actual_count) and referent ids, which differ between the two syntaxes ([MS-RPCE]
// section 2.2.5). NDR20 is the zero value, so an Encoder/Decoder built without an
// explicit syntax behaves as the classic NDR 2.0 codec.
type Syntax int

const (
	// NDR20 is NDR transfer syntax version 2.0: 4-octet counts and referent ids.
	NDR20 Syntax = iota
	// NDR64 is the NDR64 transfer syntax: 8-octet counts and referent ids.
	NDR64
)

// String returns the syntax name for diagnostics.
func (s Syntax) String() string {
	if s == NDR64 {
		return "NDR64"
	}
	return "NDR20"
}

// Encoder builds an NDR octet stream. Alignment is computed relative to the start of
// the stream, as required by [C706] section 14.2.2.
type Encoder struct {
	buf     []byte
	nextRef uint64
	syntax  Syntax
}

// firstReferentID is the referent id assigned to the first non-null pointer. The
// exact value is irrelevant on the wire (a receiver only checks non-null for
// unique/full pointers); an arbitrary non-zero base matching common implementations
// is used.
const firstReferentID uint64 = 0x00020000

// NewEncoder returns an empty encoder for the NDR20 transfer syntax.
func NewEncoder() *Encoder { return &Encoder{nextRef: firstReferentID} }

// NewEncoderForSyntax returns an empty encoder for the given transfer syntax.
func NewEncoderForSyntax(s Syntax) *Encoder { return &Encoder{nextRef: firstReferentID, syntax: s} }

// Syntax reports the transfer syntax the encoder operates under.
func (e *Encoder) Syntax() Syntax { return e.syntax }

// Bytes returns the accumulated octet stream.
func (e *Encoder) Bytes() []byte { return e.buf }

// Len returns the current stream length, which is the offset alignment is measured
// against.
func (e *Encoder) Len() int { return len(e.buf) }

// Align pads the stream with zero octets so the next write begins on a multiple of n.
func (e *Encoder) Align(n int) {
	if n <= 1 {
		return
	}
	for len(e.buf)%n != 0 {
		e.buf = append(e.buf, 0)
	}
}

// WriteBytes appends raw octets with no alignment.
func (e *Encoder) WriteBytes(b []byte) { e.buf = append(e.buf, b...) }

// WriteUint8 appends a 1-octet value.
func (e *Encoder) WriteUint8(v uint8) { e.buf = append(e.buf, v) }

// WriteUint16 appends a 2-octet little-endian value, 2-aligned.
func (e *Encoder) WriteUint16(v uint16) {
	e.Align(2)
	e.buf = binary.LittleEndian.AppendUint16(e.buf, v)
}

// WriteUint32 appends a 4-octet little-endian value, 4-aligned.
func (e *Encoder) WriteUint32(v uint32) {
	e.Align(4)
	e.buf = binary.LittleEndian.AppendUint32(e.buf, v)
}

// WriteUint64 appends an 8-octet little-endian value, 8-aligned.
func (e *Encoder) WriteUint64(v uint64) {
	e.Align(8)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, v)
}

// writeCount writes an NDR count or length determinant — a maximum_count, offset, or
// actual_count — as a 4-octet value 4-aligned under NDR20, or an 8-octet value
// 8-aligned under NDR64 ([MS-RPCE] section 2.2.5). Counts are non-negative, so a uint64
// carries either width.
func (e *Encoder) writeCount(v uint64) {
	if e.syntax == NDR64 {
		e.WriteUint64(v)
		return
	}
	e.WriteUint32(uint32(v))
}

// writeReferent writes a pointer referent id at the syntax's referent width: 4 octets
// under NDR20, 8 octets under NDR64.
func (e *Encoder) writeReferent(id uint64) {
	if e.syntax == NDR64 {
		e.WriteUint64(id)
		return
	}
	e.WriteUint32(uint32(id))
}

// nextReferent returns the next referent id for a non-null pointer, advancing by the
// syntax's referent width.
func (e *Encoder) nextReferent() uint64 {
	id := e.nextRef
	if e.syntax == NDR64 {
		e.nextRef += 8
	} else {
		e.nextRef += 4
	}
	return id
}

// writeWString writes a wchar_t string ([string] wide) as a conformant+varying array
// of UTF-16LE code units. The maximum and actual counts include the NUL terminator,
// per [MS-RPCE] Conformant Varying Strings:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/db6a89db-2c88-4ae7-90de-fca23929abab
func (e *Encoder) writeWString(s string) {
	units := utf16.EncodeUTF16LE(s) // 2 octets per code unit, no terminator
	count := uint64(len(units)/2) + 1
	e.writeCount(count) // maximum_count
	e.writeCount(0)     // offset
	e.writeCount(count) // actual_count
	e.WriteBytes(units)
	e.WriteUint16(0) // NUL terminator
}

// writeAString writes a char string ([string]) as a conformant+varying array of
// octets. The counts include the NUL terminator.
func (e *Encoder) writeAString(s string) {
	count := uint64(len(s)) + 1
	e.writeCount(count) // maximum_count
	e.writeCount(0)     // offset
	e.writeCount(count) // actual_count
	e.WriteBytes([]byte(s))
	e.WriteUint8(0) // NUL terminator
}

// writeConformantBytes writes a [size_is] byte array: a maximum_count followed by the
// elements ([MS-RPCE] Conformant Arrays:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/140b01a3-979b-43af-b1e3-28f248db8f03).
func (e *Encoder) writeConformantBytes(b []byte) {
	e.writeCount(uint64(len(b)))
	e.WriteBytes(b)
}

// Decoder reads an NDR octet stream produced under the little-endian data
// representation.
type Decoder struct {
	data   []byte
	pos    int
	syntax Syntax
}

// NewDecoder returns a decoder over data for the NDR20 transfer syntax.
func NewDecoder(data []byte) *Decoder { return &Decoder{data: data} }

// NewDecoderForSyntax returns a decoder over data for the given transfer syntax.
func NewDecoderForSyntax(data []byte, s Syntax) *Decoder { return &Decoder{data: data, syntax: s} }

// Syntax reports the transfer syntax the decoder operates under.
func (d *Decoder) Syntax() Syntax { return d.syntax }

// Pos returns the current read offset.
func (d *Decoder) Pos() int { return d.pos }

// Remaining returns the number of unread octets.
func (d *Decoder) Remaining() int { return len(d.data) - d.pos }

// Align advances the read position to the next multiple of n.
func (d *Decoder) Align(n int) {
	if n <= 1 {
		return
	}
	for d.pos%n != 0 {
		d.pos++
	}
}

// ReadBytes returns the next n octets.
func (d *Decoder) ReadBytes(n int) ([]byte, error) {
	if n < 0 || d.pos+n > len(d.data) {
		return nil, fmt.Errorf("ndr: read of %d bytes at offset %d exceeds stream length %d", n, d.pos, len(d.data))
	}
	b := d.data[d.pos : d.pos+n]
	d.pos += n
	return b, nil
}

// ReadUint8 reads a 1-octet value.
func (d *Decoder) ReadUint8() (uint8, error) {
	b, err := d.ReadBytes(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// ReadUint16 reads a 2-octet little-endian value, 2-aligned.
func (d *Decoder) ReadUint16() (uint16, error) {
	d.Align(2)
	b, err := d.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

// ReadUint32 reads a 4-octet little-endian value, 4-aligned.
func (d *Decoder) ReadUint32() (uint32, error) {
	d.Align(4)
	b, err := d.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

// ReadUint64 reads an 8-octet little-endian value, 8-aligned.
func (d *Decoder) ReadUint64() (uint64, error) {
	d.Align(8)
	b, err := d.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

// readCount reads an NDR count or length determinant at the syntax's width: 4 octets
// 4-aligned under NDR20, or 8 octets 8-aligned under NDR64 ([MS-RPCE] section 2.2.5).
func (d *Decoder) readCount() (uint64, error) {
	if d.syntax == NDR64 {
		return d.ReadUint64()
	}
	v, err := d.ReadUint32()
	return uint64(v), err
}

// readReferent reads a pointer referent id at the syntax's referent width: 4 octets
// under NDR20, 8 octets under NDR64.
func (d *Decoder) readReferent() (uint64, error) {
	if d.syntax == NDR64 {
		return d.ReadUint64()
	}
	v, err := d.ReadUint32()
	return uint64(v), err
}

// readWString reads a conformant+varying UTF-16LE string and returns it with the NUL
// terminator removed.
func (d *Decoder) readWString() (string, error) {
	if _, err := d.readCount(); err != nil { // maximum_count
		return "", err
	}
	if _, err := d.readCount(); err != nil { // offset
		return "", err
	}
	actual, err := d.readCount() // actual_count (code units, incl. terminator)
	if err != nil {
		return "", err
	}
	b, err := d.ReadBytes(int(actual) * 2)
	if err != nil {
		return "", err
	}
	return trimTerminator(utf16.DecodeUTF16LE(b)), nil
}

// readAString reads a conformant+varying ASCII string and returns it with the NUL
// terminator removed.
func (d *Decoder) readAString() (string, error) {
	if _, err := d.readCount(); err != nil { // maximum_count
		return "", err
	}
	if _, err := d.readCount(); err != nil { // offset
		return "", err
	}
	actual, err := d.readCount() // actual_count (octets, incl. terminator)
	if err != nil {
		return "", err
	}
	b, err := d.ReadBytes(int(actual))
	if err != nil {
		return "", err
	}
	return trimTerminator(string(b)), nil
}

// readConformantBytes reads a maximum_count-prefixed byte array.
func (d *Decoder) readConformantBytes() ([]byte, error) {
	n, err := d.readCount()
	if err != nil {
		return nil, err
	}
	b, err := d.ReadBytes(int(n))
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out, nil
}

// trimTerminator removes a single trailing NUL, the NDR string terminator.
func trimTerminator(s string) string {
	if n := len(s); n > 0 && s[n-1] == 0 {
		return s[:n-1]
	}
	return s
}
