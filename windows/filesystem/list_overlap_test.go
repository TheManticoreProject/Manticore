package filesystem

import (
	"encoding/binary"
	"testing"
)

// overlappingBuffer builds a 64 KiB buffer whose every 8-byte slot claims to be
// an entry with NextEntryOffset 8 and a name of 0x7FF0 bytes, the shape that
// made each entry decode most of the remaining buffer again.
func overlappingBuffer(nameLenOffset int) []byte {
	buf := make([]byte, 64*1024)
	for pos := 0; pos+nameLenOffset+4 <= len(buf); pos += 8 {
		binary.LittleEndian.PutUint32(buf[pos:], 8)
		binary.LittleEndian.PutUint32(buf[pos+nameLenOffset:], 0x7FF0)
	}
	return buf
}

// TestListParsersRefuseOverlappingEntries checks that every chained-list parser
// stops at an entry whose NextEntryOffset points inside its own fixed part and
// name, so decoded output stays bounded by the buffer.
func TestListParsersRefuseOverlappingEntries(t *testing.T) {
	cases := []struct {
		name          string
		nameLenOffset int
		parse         func([]byte) (entries, nameBytes int)
	}{
		{"FileDirectoryInformation", 60, func(b []byte) (int, int) {
			l := ParseFileDirectoryInformationList(b)
			n := 0
			for _, e := range l {
				n += len(e.FileName)
			}
			return len(l), n
		}},
		{"FileFullDirectoryInformation", 60, func(b []byte) (int, int) {
			l := ParseFileFullDirectoryInformationList(b)
			n := 0
			for _, e := range l {
				n += len(e.FileName)
			}
			return len(l), n
		}},
		{"FileBothDirectoryInformation", 60, func(b []byte) (int, int) {
			l := ParseFileBothDirectoryInformationList(b)
			n := 0
			for _, e := range l {
				n += len(e.FileName)
			}
			return len(l), n
		}},
		{"FileIdBothDirectoryInformation", 60, func(b []byte) (int, int) {
			l := ParseFileIdBothDirectoryInformationList(b)
			n := 0
			for _, e := range l {
				n += len(e.FileName)
			}
			return len(l), n
		}},
		{"FileIdExtdDirectoryInformation", 60, func(b []byte) (int, int) {
			l := ParseFileIdExtdDirectoryInformationList(b)
			n := 0
			for _, e := range l {
				n += len(e.FileName)
			}
			return len(l), n
		}},
		{"FileNamesInformation", 8, func(b []byte) (int, int) {
			l := ParseFileNamesInformationList(b)
			n := 0
			for _, e := range l {
				n += len(e.FileName)
			}
			return len(l), n
		}},
		{"FileNotifyInformation", 8, func(b []byte) (int, int) {
			l := ParseFileNotifyInformation(b)
			n := 0
			for _, e := range l {
				n += len(e.FileName)
			}
			return len(l), n
		}},
		{"FileStreamInformation", 4, func(b []byte) (int, int) {
			l := ParseFileStreamInformationList(b)
			n := 0
			for _, e := range l {
				n += len(e.StreamName)
			}
			return len(l), n
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := overlappingBuffer(tc.nameLenOffset)
			entries, nameBytes := tc.parse(buf)
			if entries != 1 {
				t.Errorf("parsed %d overlapping entries, want 1", entries)
			}
			// UTF-8 output can be up to 1.5x the UTF-16 input; anything near
			// the buffer size is the bound, not the 100 MB seen before.
			if nameBytes > 2*len(buf) {
				t.Errorf("decoded %d name bytes from a %d-byte buffer", nameBytes, len(buf))
			}
		})
	}
}

// TestListParsersFollowUnalignedChains checks that a chain whose entries are
// packed without the 8-byte padding is still read in full.
func TestListParsersFollowUnalignedChains(t *testing.T) {
	a, _ := (&FileNamesInformation{FileName: "a"}).Marshal() // 14 bytes
	b, _ := (&FileNamesInformation{FileName: "bc"}).Marshal()
	binary.LittleEndian.PutUint32(a[0:4], uint32(len(a)))
	got := ParseFileNamesInformationList(append(a, b...))
	if len(got) != 2 || got[0].FileName != "a" || got[1].FileName != "bc" {
		t.Errorf("unaligned chain parsed to %+v", got)
	}
}
