// Package ndr provides a minimal little-endian NDR (NDR 2.0) octet cursor for the
// connectionless (v4) packages that hand-marshal small, fixed wire shapes (the
// endpoint mapper and the interface bindings). It is deliberately not the
// reflection-driven codec in network/dcerpc/ndr: those callers want explicit,
// golden-testable byte layout, so this cursor just offers length-checked reads and
// writes of the NDR primitives they need, with alignment taken relative to the start
// of the stream.
//
// It lives under internal/ so it stays a v4 implementation detail rather than a
// public API alongside the full NDR package.
package ndr

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Writer accumulates a little-endian NDR octet stream. The zero value is ready to
// use.
type Writer struct {
	buf []byte
}

// U32 appends a little-endian uint32.
func (w *Writer) U32(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

// Put appends raw octets.
func (w *Writer) Put(b []byte) { w.buf = append(w.buf, b...) }

// Align pads the stream with zero octets until its length is a multiple of n.
func (w *Writer) Align(n int) {
	for len(w.buf)%n != 0 {
		w.buf = append(w.buf, 0)
	}
}

// Bytes returns the accumulated stream.
func (w *Writer) Bytes() []byte { return w.buf }

// Reader consumes a little-endian NDR octet stream with bounds checking; the first
// error is sticky and reported by Err.
type Reader struct {
	data []byte
	off  int
	fail error
}

// NewReader returns a reader over data.
func NewReader(data []byte) *Reader { return &Reader{data: data} }

// Err returns the first error encountered, or nil.
func (r *Reader) Err() error { return r.fail }

// Take returns the next n octets, or nil and a sticky error if fewer remain.
func (r *Reader) Take(n int) []byte {
	if r.fail != nil {
		return nil
	}
	if n < 0 || r.off+n > len(r.data) {
		r.fail = fmt.Errorf("ndr: underrun reading %d bytes at offset %d", n, r.off)
		return nil
	}
	b := r.data[r.off : r.off+n]
	r.off += n
	return b
}

// Skip advances past n octets.
func (r *Reader) Skip(n int) { r.Take(n) }

// U16 reads a little-endian uint16.
func (r *Reader) U16() uint16 {
	b := r.Take(2)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

// U32 reads a little-endian uint32.
func (r *Reader) U32() uint32 {
	b := r.Take(4)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}

// UUID reads a 16-octet DCE uuid_t.
func (r *Reader) UUID() guid.GUID {
	var g guid.GUID
	if b := r.Take(16); b != nil {
		g.FromRawBytes(b)
	}
	return g
}

// Align advances to the next multiple of n, relative to the start of the stream.
func (r *Reader) Align(n int) {
	if r.fail != nil {
		return
	}
	for r.off%n != 0 {
		if r.off >= len(r.data) {
			r.fail = fmt.Errorf("ndr: underrun aligning to %d at offset %d", n, r.off)
			return
		}
		r.off++
	}
}
