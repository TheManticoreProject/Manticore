// Package ndr implements the DCE/RPC Network Data Representation (NDR) transfer
// syntax, version 2.0 (the "NDR20" little-endian encoding used by Windows RPC), with
// a declarative, reflection-driven API for marshalling RPC call structures.
//
// The low-level codec (Encoder/Decoder) handles NDR's stream-relative alignment and
// primitive encoding; the reflection walker in marshal.go drives it from struct tags
// so callers can declare an RPC call as a Go struct (see Call, Marshal, Unmarshal).
//
// Scope: NDR20, little-endian (the standard Windows data representation, DREP 0x10).
// Big-endian and NDR64 (8-byte counts/referents) are out of scope.
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

// Encoder builds an NDR octet stream. Alignment is computed relative to the start of
// the stream, as required by [C706] section 14.2.2.
type Encoder struct {
	buf     []byte
	nextRef uint32
}

// firstReferentID is the referent id assigned to the first non-null pointer. The
// exact value is irrelevant on the wire (a receiver only checks non-null for
// unique/full pointers); an arbitrary non-zero base matching common implementations
// is used.
const firstReferentID uint32 = 0x00020000

// NewEncoder returns an empty encoder.
func NewEncoder() *Encoder { return &Encoder{nextRef: firstReferentID} }

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

// nextReferent returns the next referent id for a non-null pointer.
func (e *Encoder) nextReferent() uint32 {
	id := e.nextRef
	e.nextRef += 4
	return id
}

// writeWString writes a wchar_t string ([string] wide) as a conformant+varying array
// of UTF-16LE code units. The maximum and actual counts include the NUL terminator,
// per [MS-RPCE] Conformant Varying Strings:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/db6a89db-2c88-4ae7-90de-fca23929abab
func (e *Encoder) writeWString(s string) {
	units := utf16.EncodeUTF16LE(s) // 2 octets per code unit, no terminator
	count := uint32(len(units)/2) + 1
	e.WriteUint32(count) // maximum_count
	e.WriteUint32(0)     // offset
	e.WriteUint32(count) // actual_count
	e.WriteBytes(units)
	e.WriteUint16(0) // NUL terminator
}

// writeAString writes a char string ([string]) as a conformant+varying array of
// octets. The counts include the NUL terminator.
func (e *Encoder) writeAString(s string) {
	count := uint32(len(s)) + 1
	e.WriteUint32(count) // maximum_count
	e.WriteUint32(0)     // offset
	e.WriteUint32(count) // actual_count
	e.WriteBytes([]byte(s))
	e.WriteUint8(0) // NUL terminator
}

// writeConformantBytes writes a [size_is] byte array: a maximum_count followed by the
// elements ([MS-RPCE] Conformant Arrays:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/140b01a3-979b-43af-b1e3-28f248db8f03).
func (e *Encoder) writeConformantBytes(b []byte) {
	e.WriteUint32(uint32(len(b)))
	e.WriteBytes(b)
}

// Decoder reads an NDR octet stream produced under the little-endian data
// representation.
type Decoder struct {
	data []byte
	pos  int
}

// NewDecoder returns a decoder over data.
func NewDecoder(data []byte) *Decoder { return &Decoder{data: data} }

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

// readWString reads a conformant+varying UTF-16LE string and returns it with the NUL
// terminator removed.
func (d *Decoder) readWString() (string, error) {
	if _, err := d.ReadUint32(); err != nil { // maximum_count
		return "", err
	}
	if _, err := d.ReadUint32(); err != nil { // offset
		return "", err
	}
	actual, err := d.ReadUint32() // actual_count (code units, incl. terminator)
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
	if _, err := d.ReadUint32(); err != nil { // maximum_count
		return "", err
	}
	if _, err := d.ReadUint32(); err != nil { // offset
		return "", err
	}
	actual, err := d.ReadUint32() // actual_count (octets, incl. terminator)
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
	n, err := d.ReadUint32()
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
