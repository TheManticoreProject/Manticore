package ndr

import (
	"encoding/binary"
	"math"
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
