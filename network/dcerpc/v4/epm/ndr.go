package epm

import (
	"encoding/binary"
	"fmt"
)

// This file holds a minimal little-endian NDR (NDR 2.0) cursor used to (un)marshal
// the small, fixed shape of the ept_map request and response. It is deliberately
// self-contained rather than reflection-driven: the ept_map layout is tiny and
// hand-coding it keeps the byte layout explicit and golden-testable. NDR alignment
// is taken relative to the start of the stub, which is where these cursors begin.

// ndrWriter accumulates a little-endian NDR octet stream.
type ndrWriter struct {
	buf []byte
}

func (w *ndrWriter) u32(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

func (w *ndrWriter) bytes(b []byte) { w.buf = append(w.buf, b...) }

// align pads the stream with zero octets until its length is a multiple of n.
func (w *ndrWriter) align(n int) {
	for len(w.buf)%n != 0 {
		w.buf = append(w.buf, 0)
	}
}

// ndrReader consumes a little-endian NDR octet stream with bounds checking; the first
// error is sticky and reported by err().
type ndrReader struct {
	data []byte
	off  int
	fail error
}

func (r *ndrReader) err() error { return r.fail }

func (r *ndrReader) u32() uint32 {
	if r.fail != nil {
		return 0
	}
	if r.off+4 > len(r.data) {
		r.fail = fmt.Errorf("epm: NDR underrun reading uint32 at offset %d", r.off)
		return 0
	}
	v := binary.LittleEndian.Uint32(r.data[r.off:])
	r.off += 4
	return v
}

// take returns the next n octets, or nil and a sticky error if fewer remain.
func (r *ndrReader) take(n int) []byte {
	if r.fail != nil {
		return nil
	}
	if n < 0 || r.off+n > len(r.data) {
		r.fail = fmt.Errorf("epm: NDR underrun reading %d bytes at offset %d", n, r.off)
		return nil
	}
	b := r.data[r.off : r.off+n]
	r.off += n
	return b
}

// skip advances past n octets.
func (r *ndrReader) skip(n int) { r.take(n) }

// align advances to the next multiple of n, relative to the start of the stream.
func (r *ndrReader) align(n int) {
	if r.fail != nil {
		return
	}
	for r.off%n != 0 {
		if r.off >= len(r.data) {
			r.fail = fmt.Errorf("epm: NDR underrun aligning to %d at offset %d", n, r.off)
			return
		}
		r.off++
	}
}
