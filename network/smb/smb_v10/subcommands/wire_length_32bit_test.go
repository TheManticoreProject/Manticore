package subcommands

import (
	"encoding/binary"
	"testing"
)

// The tests in this file declare 32-bit lengths and entry offsets that wrap
// negative when converted to a 32-bit int. They are only meaningful under
// GOARCH=386, where the unguarded conversions passed the bounds checks and
// panicked on the slice; on a 64-bit build they still pin the result.

// oversizedLengths overflow offset+length in a 32-bit int, or are negative as one.
var oversizedLengths = []uint32{0x7FFFFFF8, 0x80000000, 0xFFFFFFF0}

// withUint32 returns a zeroed buffer of size bytes with v written at at.
func withUint32(size, at int, v uint32) []byte {
	b := make([]byte, size)
	binary.LittleEndian.PutUint32(b[at:], v)
	return b
}

// decodeWithoutPanic runs decode and fails the test, rather than the process, if it panics.
func decodeWithoutPanic(t *testing.T, name string, decode func() error) error {
	t.Helper()
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("%s panicked: %v", name, r)
			}
		}()
		err = decode()
	}()
	return err
}

func TestVariableLengthOverflowRejected(t *testing.T) {
	for _, length := range oversizedLengths {
		cases := []struct {
			name   string
			decode func() error
		}{
			{"FileNotifyInformation.Unmarshal", func() error {
				_, err := (&FileNotifyInformation{}).Unmarshal(withUint32(16, 8, length))
				return err
			}},
			{"FileGetQuotaInformation.Unmarshal", func() error {
				_, err := (&FileGetQuotaInformation{}).Unmarshal(withUint32(16, 4, length))
				return err
			}},
			{"FileQuotaInformation.Unmarshal", func() error {
				_, err := (&FileQuotaInformation{}).Unmarshal(withUint32(48, 4, length))
				return err
			}},
			{"SrvRequestResumeKeyResponse.Unmarshal", func() error {
				_, err := (&SrvRequestResumeKeyResponse{}).Unmarshal(withUint32(CopychunkResumeKeyLength+8, CopychunkResumeKeyLength, length))
				return err
			}},
		}
		for _, c := range cases {
			if err := decodeWithoutPanic(t, c.name, c.decode); err == nil {
				t.Errorf("%s: length 0x%08X: accepted a field past the end of the data", c.name, length)
			}
		}
	}
}

func TestSrvSnapshotArrayOversizedSizeClamped(t *testing.T) {
	for _, size := range oversizedLengths {
		data := withUint32(16, 8, size)
		var a SrvSnapshotArray
		var consumed int
		err := decodeWithoutPanic(t, "SrvSnapshotArray.Unmarshal", func() error {
			var err error
			consumed, err = a.Unmarshal(data)
			return err
		})
		// The list is read only as far as the data goes, as for a buffer probe.
		if err != nil {
			t.Errorf("size 0x%08X: Unmarshal: %v", size, err)
		}
		if consumed > len(data) {
			t.Errorf("size 0x%08X: consumed %d of %d bytes", size, consumed, len(data))
		}
	}
}

func TestNextEntryOffsetOutOfRangeRejected(t *testing.T) {
	for _, next := range []uint32{4, 0x7FFFFFF8, 0x80000000, 0xFFFFFFF0} {
		err := decodeWithoutPanic(t, "ParseFileNotifyInformationList", func() error {
			_, err := ParseFileNotifyInformationList(withUint32(32, 0, next))
			return err
		})
		if err == nil {
			t.Errorf("FILE_NOTIFY_INFORMATION NextEntryOffset 0x%08X: accepted a link outside the list", next)
		}
		err = decodeWithoutPanic(t, "ParseFileQuotaInformationList", func() error {
			_, err := ParseFileQuotaInformationList(withUint32(96, 0, next))
			return err
		})
		if err == nil {
			t.Errorf("FILE_QUOTA_INFORMATION NextEntryOffset 0x%08X: accepted a link outside the list", next)
		}
	}
}
