package client

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// bothDirEntry builds a single FILE_BOTH_DIR_INFORMATION entry with a UTF-16LE
// name. next is written to NextEntryOffset verbatim.
func bothDirEntry(name string, attrs uint32, size uint64, next uint32) []byte {
	u16 := utf16.Encode([]rune(name))
	nameBytes := make([]byte, len(u16)*2)
	for i, u := range u16 {
		binary.LittleEndian.PutUint16(nameBytes[i*2:], u)
	}
	e := make([]byte, bothDirInfoFixedSize+len(nameBytes))
	binary.LittleEndian.PutUint32(e[0:4], next)
	binary.LittleEndian.PutUint64(e[40:48], size)                   // EndOfFile
	binary.LittleEndian.PutUint64(e[48:56], size)                   // AllocationSize
	binary.LittleEndian.PutUint32(e[56:60], attrs)                  // FileAttributes
	binary.LittleEndian.PutUint32(e[60:64], uint32(len(nameBytes))) // FileNameLength
	copy(e[bothDirInfoFixedSize:], nameBytes)
	return e
}

func TestParseBothDirectoryInfo(t *testing.T) {
	dot := bothDirEntry(".", fileAttributeDirectory, 0, 0)
	dotdot := bothDirEntry("..", fileAttributeDirectory, 0, 0)
	file := bothDirEntry("report.txt", 0x20 /*ARCHIVE*/, 1234, 0)
	dir := bothDirEntry("subdir", fileAttributeDirectory, 0, 0)

	// Chain: NextEntryOffset of each non-last entry is that entry's length.
	dot[0] = byte(len(dot))
	off := len(dot)
	binary.LittleEndian.PutUint32(dotdot[0:4], uint32(len(dotdot)))
	off += len(dotdot)
	binary.LittleEndian.PutUint32(file[0:4], uint32(len(file)))
	off += len(file)
	_ = off

	var buf []byte
	buf = append(buf, dot...)
	buf = append(buf, dotdot...)
	buf = append(buf, file...)
	buf = append(buf, dir...) // last, next=0

	got := parseBothDirectoryInfo(buf)

	// "." and ".." are skipped.
	if len(got) != 2 {
		t.Fatalf("expected 2 entries (. and .. skipped), got %d: %+v", len(got), got)
	}
	if got[0].Name != "report.txt" || got[0].IsDir() || got[0].Size != 1234 {
		t.Errorf("file entry wrong: %+v", got[0])
	}
	if got[1].Name != "subdir" || !got[1].IsDir() {
		t.Errorf("dir entry wrong: %+v", got[1])
	}
}

func TestParseBothDirectoryInfo_Empty(t *testing.T) {
	if got := parseBothDirectoryInfo(nil); got != nil {
		t.Errorf("expected nil for empty buffer, got %+v", got)
	}
	// A truncated buffer (shorter than the fixed entry) yields no entries.
	if got := parseBothDirectoryInfo(make([]byte, 10)); len(got) != 0 {
		t.Errorf("expected no entries for truncated buffer, got %d", len(got))
	}
}
