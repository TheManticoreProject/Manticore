package filesystem

import (
	"encoding/binary"
	"testing"
)

// hugeLengths are 32-bit wire lengths/offsets of 2 GiB and above. On 32-bit
// platforms int(v) wraps them negative, or adding them to a small position
// overflows int, which used to defeat the upper-bound checks below.
var hugeLengths = []uint32{0x7FFFFFF0, 0x7FFFFFFF, 0x80000000, 0xFFFFFFF0}

// TestUnmarshal_HugeNameLength checks that a FileNameLength (or label length)
// of 2 GiB or more leaves the name empty instead of slicing out of range.
func TestUnmarshal_HugeNameLength(t *testing.T) {
	cases := []struct {
		name      string
		size      int
		lenPos    int
		unmarshal func([]byte) (string, error)
	}{
		{"FileFullDirectoryInformation", fileFullDirInfoFixedSize, 60, func(b []byte) (string, error) {
			var fi FileFullDirectoryInformation
			err := fi.Unmarshal(b)
			return fi.FileName, err
		}},
		{"FileIdFullDirectoryInformation", fileIdFullDirInfoFixedSize, 60, func(b []byte) (string, error) {
			var fi FileIdFullDirectoryInformation
			err := fi.Unmarshal(b)
			return fi.FileName, err
		}},
		{"FileBothDirectoryInformation", fileBothDirInfoFixedSize, 60, func(b []byte) (string, error) {
			var fi FileBothDirectoryInformation
			err := fi.Unmarshal(b)
			return fi.FileName, err
		}},
		{"FileIdBothDirectoryInformation", fileIdBothDirInfoFixedSize, 60, func(b []byte) (string, error) {
			var fi FileIdBothDirectoryInformation
			err := fi.Unmarshal(b)
			return fi.FileName, err
		}},
		{"FileNamesInformation", fileNamesInfoFixedSize, 8, func(b []byte) (string, error) {
			var fi FileNamesInformation
			err := fi.Unmarshal(b)
			return fi.FileName, err
		}},
		{"FileAllInformation", fileAllInformationFixedSize, fileAllInformationFixedSize - 4, func(b []byte) (string, error) {
			var fi FileAllInformation
			err := fi.Unmarshal(b)
			return fi.FileName, err
		}},
		{"FileFsVolumeInformation", fileFsVolumeInfoFixedSize, 12, func(b []byte) (string, error) {
			var fi FileFsVolumeInformation
			err := fi.Unmarshal(b)
			return fi.VolumeLabel, err
		}},
		{"FileFsAttributeInformation", fileFsAttributeInfoFixedSize, 8, func(b []byte) (string, error) {
			var fi FileFsAttributeInformation
			err := fi.Unmarshal(b)
			return fi.FileSystemName, err
		}},
	}
	for _, c := range cases {
		for _, n := range hugeLengths {
			data := make([]byte, c.size+32)
			binary.LittleEndian.PutUint32(data[c.lenPos:], n)
			name, err := c.unmarshal(data)
			if err != nil {
				t.Errorf("%s: name length %#x: unexpected error %v", c.name, n, err)
			}
			if name != "" {
				t.Errorf("%s: name length %#x: name = %q, want empty", c.name, n, name)
			}
		}
	}
}

// TestFileRenameInformation_HugeNameLength checks that a FileNameLength of
// 2 GiB or more is rejected with an error instead of panicking.
func TestFileRenameInformation_HugeNameLength(t *testing.T) {
	for _, n := range hugeLengths {
		data := make([]byte, 24)
		binary.LittleEndian.PutUint32(data[16:20], n)
		var fi FileRenameInformation
		if err := fi.Unmarshal(data); err == nil {
			t.Errorf("name length %#x: Unmarshal accepted an out-of-bounds length, want error", n)
		}
	}
}

// TestParseLists_HugeNextEntryOffset checks that a NextEntryOffset of 2 GiB or
// more ends the chain after the first entry instead of moving the read
// position out of the buffer.
func TestParseLists_HugeNextEntryOffset(t *testing.T) {
	first, _ := (&FileNamesInformation{FileName: "a"}).Marshal()
	for _, n := range hugeLengths {
		buf := append(append([]byte{}, first...), make([]byte, 128)...)
		binary.LittleEndian.PutUint32(buf[0:4], n)

		if got := ParseFileNamesInformationList(buf); len(got) != 1 || got[0].FileName != "a" {
			t.Errorf("FileNames NextEntryOffset %#x: got %+v, want the single entry \"a\"", n, got)
		}
		if got := ParseFileFullDirectoryInformationList(buf); len(got) != 1 {
			t.Errorf("FileFullDir NextEntryOffset %#x: got %d entries, want 1", n, len(got))
		}
		if got := ParseFileBothDirectoryInformationList(buf); len(got) != 1 {
			t.Errorf("FileBothDir NextEntryOffset %#x: got %d entries, want 1", n, len(got))
		}
		if got := ParseFileIdBothDirectoryInformationList(buf); len(got) != 1 {
			t.Errorf("FileIdBothDir NextEntryOffset %#x: got %d entries, want 1", n, len(got))
		}
		if got := ParseFileNotifyInformation(buf); len(got) != 1 {
			t.Errorf("FileNotify NextEntryOffset %#x: got %d records, want 1", n, len(got))
		}
	}
}
