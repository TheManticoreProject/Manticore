package client

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// bothDirEntry builds a single FILE_BOTH_DIR_INFORMATION entry via the typed
// windows/filesystem marshaller. next is written to NextEntryOffset verbatim.
func bothDirEntry(name string, attrs uint32, size uint64, next uint32) []byte {
	e := &filesystem.FileBothDirectoryInformation{
		NextEntryOffset: next,
		EndOfFile:       int64(size),
		AllocationSize:  int64(size),
		FileAttributes:  attrs,
		FileName:        name,
	}
	b, err := e.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}

func TestParseBothDirectoryInfo(t *testing.T) {
	dot := bothDirEntry(".", fileflags.FILE_ATTRIBUTE_DIRECTORY, 0, 0)
	dotdot := bothDirEntry("..", fileflags.FILE_ATTRIBUTE_DIRECTORY, 0, 0)
	file := bothDirEntry("report.txt", 0x20 /*ARCHIVE*/, 1234, 0)
	dir := bothDirEntry("subdir", fileflags.FILE_ATTRIBUTE_DIRECTORY, 0, 0)

	// Chain: NextEntryOffset of each non-last entry is that entry's length.
	binary.LittleEndian.PutUint32(dot[0:4], uint32(len(dot)))
	binary.LittleEndian.PutUint32(dotdot[0:4], uint32(len(dotdot)))
	binary.LittleEndian.PutUint32(file[0:4], uint32(len(file)))

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
