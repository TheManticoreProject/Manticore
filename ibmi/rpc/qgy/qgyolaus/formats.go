package qgyolaus

import (
	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyolaus/structures"
)

// Format is a sealed QGYOLAUS receiver-format descriptor. It couples the
// eight-character format name, the exact record length and the decoder at
// compile time, so a caller cannot send "AUTU0150" and decode with another
// layout. Obtain one from AUTU0100, AUTU0150, AUTU0200 or AUTU0250; the
// zero Format is rejected.
type Format[T any] struct {
	name   string
	length uint32
	decode func([]byte, ccsid.Codec) (T, error)
}

// AUTU0100 selects the 12-byte format: name and indicators.
func AUTU0100() Format[structures.AUTU0100] {
	return Format[structures.AUTU0100]{"AUTU0100", structures.AUTU0100Length, structures.DecodeAUTU0100}
}

// AUTU0150 selects the 62-byte format: AUTU0100 plus text description.
func AUTU0150() Format[structures.AUTU0150] {
	return Format[structures.AUTU0150]{"AUTU0150", structures.AUTU0150Length, structures.DecodeAUTU0150}
}

// AUTU0200 selects the 176-byte format: AUTU0100 plus group profiles.
func AUTU0200() Format[structures.AUTU0200] {
	return Format[structures.AUTU0200]{"AUTU0200", structures.AUTU0200Length, structures.DecodeAUTU0200}
}

// AUTU0250 selects the 228-byte format: description and group profiles.
func AUTU0250() Format[structures.AUTU0250] {
	return Format[structures.AUTU0250]{"AUTU0250", structures.AUTU0250Length, structures.DecodeAUTU0250}
}

// Name returns the eight-character format name.
func (f Format[T]) Name() string { return f.name }

// RecordLength returns the exact record length.
func (f Format[T]) RecordLength() uint32 { return f.length }

// IsZero reports whether f is the unusable zero descriptor.
func (f Format[T]) IsZero() bool { return f.decode == nil }

// DecodeRecords decodes count consecutive records from the start of
// receiver. The receiver must hold at least count records.
func (f Format[T]) DecodeRecords(receiver []byte, count int, codec ccsid.Codec) ([]T, error) {
	if f.IsZero() {
		return nil, errZeroFormat
	}
	if count < 0 || uint64(count)*uint64(f.length) > uint64(len(receiver)) {
		return nil, errRecords(count, f.length, len(receiver))
	}
	out := make([]T, 0, count)
	for i := 0; i < count; i++ {
		off := i * int(f.length)
		r, err := f.decode(receiver[off:off+int(f.length)], codec)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}
