package ndr

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"testing"
)

// mustNotPanic runs fn and fails the test if it panics.
func mustNotPanic(t *testing.T, fn func() error) error {
	t.Helper()
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("decoder panicked on a hostile count: %v", r)
			}
		}()
		err = fn()
	}()
	return err
}

// TestDecoder_ReadBytesHugeLengthRejected verifies a length near MaxInt is rejected
// instead of overflowing pos+n past the bounds check and panicking on the slice.
func TestDecoder_ReadBytesHugeLengthRejected(t *testing.T) {
	d := NewDecoder([]byte{1, 2, 3, 4})
	if _, err := d.ReadBytes(1); err != nil {
		t.Fatalf("ReadBytes(1): %v", err)
	}
	err := mustNotPanic(t, func() error {
		_, err := d.ReadBytes(math.MaxInt)
		return err
	})
	if err == nil {
		t.Fatal("ReadBytes(MaxInt): error = nil, want non-nil")
	}
}

// TestUnmarshal_NDR64HugeStringCountRejected verifies an NDR64 string whose 8-octet
// actual_count overflows when scaled to bytes is rejected without panicking.
func TestUnmarshal_NDR64HugeStringCountRejected(t *testing.T) {
	for _, tc := range []struct {
		name   string
		actual uint64
		decode func([]byte) error
	}{
		{"ascii", 0x3FFFFFFFFFFFFFF8, func(b []byte) error {
			var s struct {
				Name string `ndr:"ascii"`
			}
			return UnmarshalAs(b, &s, NDR64)
		}},
		{"wstring", 0x3FFFFFFFFFFFFFF8, func(b []byte) error {
			var s struct{ Name string }
			return UnmarshalAs(b, &s, NDR64)
		}},
		{"wstring-wraps-to-small", 0x8000000000000002, func(b []byte) error {
			var s struct{ Name string }
			return UnmarshalAs(b, &s, NDR64)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := make([]byte, 32)
			binary.LittleEndian.PutUint64(buf[0:], 1)          // maximum_count
			binary.LittleEndian.PutUint64(buf[8:], 0)          // offset
			binary.LittleEndian.PutUint64(buf[16:], tc.actual) // actual_count
			if err := mustNotPanic(t, func() error { return tc.decode(buf) }); err == nil {
				t.Fatalf("actual_count %#x: error = nil, want non-nil", tc.actual)
			}
		})
	}
}

// TestDecoder_HugeConformantCountRejected covers the NDR20 counts that overflow
// pos+n (or count*2) on 32-bit platforms.
func TestDecoder_HugeConformantCountRejected(t *testing.T) {
	t.Run("conformant-bytes", func(t *testing.T) {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint32(buf, 0x7FFFFFFE)
		d := NewDecoder(buf)
		if err := mustNotPanic(t, func() error { _, err := d.readConformantBytes(); return err }); err == nil {
			t.Fatal("readConformantBytes: error = nil, want non-nil")
		}
	})
	t.Run("wstring", func(t *testing.T) {
		buf := make([]byte, 12)
		binary.LittleEndian.PutUint32(buf[0:], 1)
		binary.LittleEndian.PutUint32(buf[8:], 0x3FFFFFF8)
		var s struct{ Name string }
		if err := mustNotPanic(t, func() error { return Unmarshal(buf, &s) }); err == nil {
			t.Fatal("Unmarshal: error = nil, want non-nil")
		}
	})
}

// TestUnmarshal_NDR64ArrayCountNotTruncated verifies an NDR64 element count whose high
// 32 bits are set is rejected rather than truncated to its low 32 bits when converted
// to int on a 32-bit platform (GOARCH=386), and that the error reports the wire count
// (not a value wrapped negative) for counts at or above 2^63. Each stub is a valid
// 2-element encoding whose count field is overwritten with the hostile value.
func TestUnmarshal_NDR64ArrayCountNotTruncated(t *testing.T) {
	type conformantPtr struct {
		P []uint8 `ndr:"unique,conformant"`
	}
	type varyingPtr struct {
		P []uint16 `ndr:"unique,varying"`
	}
	type conformantEmbedded struct {
		N uint32
		P []uint8 `ndr:"conformant"`
	}
	type varyingEmbedded struct {
		N uint32
		P []uint16 `ndr:"varying"`
	}
	for _, tc := range []struct {
		name   string
		stub   []byte
		offset int // position of the 8-octet count to overwrite
		decode func([]byte) error
	}{
		{"conformant", []byte{
			0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // referent id
			0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // maximum_count
			0xaa, 0xbb,
		}, 8, func(b []byte) error { var v conformantPtr; return UnmarshalAs(b, &v, NDR64) }},
		{"varying", []byte{
			0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // referent id
			0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // maximum_count
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // offset
			0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // actual_count
			0x00, 0x00, 0x01, 0x00,
		}, 24, func(b []byte) error { var v varyingPtr; return UnmarshalAs(b, &v, NDR64) }},
		{"embedded-conformant", []byte{
			0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // maximum_count (hoisted)
			0x07, 0x00, 0x00, 0x00, // N
			0xaa, 0xbb,
		}, 0, func(b []byte) error { var v conformantEmbedded; return UnmarshalAs(b, &v, NDR64) }},
		{"embedded-varying", []byte{
			0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // maximum_count (hoisted)
			0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // N + padding
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // offset
			0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // actual_count
			0x00, 0x00, 0x01, 0x00,
		}, 24, func(b []byte) error { var v varyingEmbedded; return UnmarshalAs(b, &v, NDR64) }},
		{"pipe", []byte{
			0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // chunk count
			0xaa, 0xbb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // elements + padding
			0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // -count trailer
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // terminator
			0x00, 0x00, 0x00, 0x00, // retval
		}, 0, func(b []byte) error { var v pipeResp; return UnmarshalAs(b, &v, NDR64) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.decode(tc.stub); err != nil {
				t.Fatalf("valid stub: %v", err)
			}
			for _, count := range []uint64{0x100000002, 0x8000000000000002} {
				buf := append([]byte(nil), tc.stub...)
				binary.LittleEndian.PutUint64(buf[tc.offset:], count)
				err := mustNotPanic(t, func() error { return tc.decode(buf) })
				if err == nil {
					t.Fatalf("count %#x: error = nil, want non-nil", count)
				}
				if want := fmt.Sprintf("%d", count); !strings.Contains(err.Error(), want) {
					t.Errorf("count %#x: error %q does not report the wire count %s", count, err, want)
				}
			}
		})
	}
}
