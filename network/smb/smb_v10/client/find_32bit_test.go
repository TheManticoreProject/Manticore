package client

import (
	"encoding/binary"
	"testing"
)

// TestParseBothDirInfoOversizedFieldsStop asserts a FileNameLength or
// NextEntryOffset that does not fit the buffer is not followed. As an int, a
// value from 0x80000000 is negative on a 32-bit build and pos+length wraps from
// 0x7FFFFFFF, which passed the bounds checks and panicked on the slice; the test
// is only meaningful under GOARCH=386.
func TestParseBothDirInfoOversizedFieldsStop(t *testing.T) {
	for _, tc := range []struct {
		name  string
		at    int
		value uint32
	}{
		{"NextEntryOffset", 0, 0x7FFFFFF8},
		{"NextEntryOffset", 0, 0x80000000},
		{"NextEntryOffset", 0, 0xFFFFFFF0},
		{"NextEntryOffset", 0, 4},
		{"FileNameLength", 60, 0x7FFFFFF0},
		{"FileNameLength", 60, 0x80000000},
		{"FileNameLength", 60, 0xFFFFFFF0},
	} {
		first := makeBothDirInfoEntry("ONE.TXT", "", 0, 0, 0, 0, 0, 0, uint32(bothDirInfoFixedSize+len("ONE.TXT")))
		second := makeBothDirInfoEntry("TWO.TXT", "", 0, 0, 0, 0, 0, 0, 0)
		data := append(first, second...)
		binary.LittleEndian.PutUint32(data[tc.at:], tc.value)

		var entries []Entry
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s 0x%08X: parseBothDirInfo panicked: %v", tc.name, tc.value, r)
				}
			}()
			entries = parseBothDirInfo(data, false)
		}()

		switch tc.name {
		case "NextEntryOffset":
			// The first entry is intact; the link out of it is not followed.
			if len(entries) != 1 || entries[0].LongName != "ONE.TXT" {
				t.Errorf("NextEntryOffset 0x%08X: got %+v, want only ONE.TXT", tc.value, entries)
			}
		case "FileNameLength":
			// The name does not fit, so none is decoded; the next entry still is.
			if len(entries) != 2 || entries[0].LongName != "" || entries[1].LongName != "TWO.TXT" {
				t.Errorf("FileNameLength 0x%08X: got %+v, want an unnamed entry then TWO.TXT", tc.value, entries)
			}
		}
	}
}
