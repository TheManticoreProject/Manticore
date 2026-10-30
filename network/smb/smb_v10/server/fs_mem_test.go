package server

import (
	"errors"
	"math"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// TestMemoryFileSystemRefusesSizeBeyondVolume checks that a write, a truncate or
// a size change that would grow a file past the volume's capacity fails with
// ErrDiskFull instead of allocating it.
func TestMemoryFileSystemRefusesSizeBeyondVolume(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("f.bin", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	file, err := fs.Open("f.bin", OpenFlags{Read: true, Write: true})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	limit := fs.volume.TotalBytes
	for _, off := range []int64{1 << 40, limit, limit - 1, math.MaxInt64 - 1, math.MaxInt64} {
		if n, err := file.WriteAt([]byte("xy"), off); !errors.Is(err, ErrDiskFull) {
			t.Errorf("WriteAt(off=%d) = %d, %v; want ErrDiskFull", off, n, err)
		}
	}
	for _, size := range []int64{1 << 40, limit + 1, math.MaxInt64} {
		if err := file.Truncate(size); !errors.Is(err, ErrDiskFull) {
			t.Errorf("Truncate(%d) = %v; want ErrDiskFull", size, err)
		}
		if err := fs.SetAttr("f.bin", FileAttr{Size: size}, AttrMask{Size: true}); !errors.Is(err, ErrDiskFull) {
			t.Errorf("SetAttr(Size=%d) = %v; want ErrDiskFull", size, err)
		}
	}

	// The refused calls left the file as it was.
	if attr, err := fs.Stat("f.bin"); err != nil || attr.Size != 5 {
		t.Errorf("Stat = %+v, %v; want the original 5 bytes", attr, err)
	}
	// Sizes within the volume still work.
	if _, err := file.WriteAt([]byte("!"), 4095); err != nil {
		t.Errorf("WriteAt within the volume: %v", err)
	}
	if err := file.Truncate(10); err != nil {
		t.Errorf("Truncate within the volume: %v", err)
	}
	if got := statusForFSError(ErrDiskFull); got != nt_status.NT_STATUS_DISK_FULL {
		t.Errorf("statusForFSError(ErrDiskFull) = %v, want NT_STATUS_DISK_FULL", got)
	}
}

// TestMemoryFileSystemCapacityIsShared checks that the volume's capacity bounds
// all files together, not each one: two files whose sizes add up to more than
// the volume are refused, the free space reported follows the usage, and
// removing, truncating or replacing a file gives its space back.
func TestMemoryFileSystemCapacityIsShared(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	fs.volume.TotalBytes = 1000

	free := func() int64 {
		t.Helper()
		volume, err := fs.VolumeInfo()
		if err != nil {
			t.Fatal(err)
		}
		return volume.FreeBytes
	}
	open := func(path string) File {
		t.Helper()
		file, err := fs.Open(path, OpenFlags{Read: true, Write: true, Create: true})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { file.Close() })
		return file
	}

	if err := fs.AddFile("seed.txt", make([]byte, 100)); err != nil {
		t.Fatal(err)
	}
	if got := free(); got != 900 {
		t.Fatalf("FreeBytes after a 100-byte file = %d, want 900", got)
	}

	a, b := open("a.bin"), open("b.bin")
	if err := a.Truncate(600); err != nil {
		t.Fatalf("Truncate(a, 600): %v", err)
	}
	// 100 + 600 + 301 is past the volume, whichever way b grows.
	if err := b.Truncate(301); !errors.Is(err, ErrDiskFull) {
		t.Errorf("Truncate(b, 301) = %v, want ErrDiskFull", err)
	}
	if _, err := b.WriteAt([]byte("x"), 300); !errors.Is(err, ErrDiskFull) {
		t.Errorf("WriteAt(b, off=300) = %v, want ErrDiskFull", err)
	}
	if err := fs.SetAttr("b.bin", FileAttr{Size: 301}, AttrMask{Size: true}); !errors.Is(err, ErrDiskFull) {
		t.Errorf("SetAttr(b, Size=301) = %v, want ErrDiskFull", err)
	}
	if err := fs.AddFile("c.bin", make([]byte, 301)); !errors.Is(err, ErrDiskFull) {
		t.Errorf("AddFile(c, 301 bytes) = %v, want ErrDiskFull", err)
	}
	// Exactly filling the volume is allowed.
	if _, err := b.WriteAt([]byte("x"), 299); err != nil {
		t.Errorf("WriteAt(b, off=299): %v", err)
	}
	if got := free(); got != 0 {
		t.Errorf("FreeBytes on a full volume = %d, want 0", got)
	}

	// A hard link shares the data: removing one name frees nothing.
	if err := fs.Link("a.bin", "a-link.bin"); err != nil {
		t.Fatal(err)
	}
	if err := fs.Remove("a.bin"); err != nil {
		t.Fatal(err)
	}
	if got := free(); got != 0 {
		t.Errorf("FreeBytes after removing one of two names = %d, want 0", got)
	}
	if err := fs.Remove("a-link.bin"); err != nil {
		t.Fatal(err)
	}
	if got := free(); got != 600 {
		t.Errorf("FreeBytes after removing the last name = %d, want 600", got)
	}

	// Shrinking, replacing and truncating on open give space back too.
	if err := b.Truncate(100); err != nil {
		t.Fatal(err)
	}
	if got := free(); got != 800 {
		t.Errorf("FreeBytes after shrinking b to 100 = %d, want 800", got)
	}
	if err := fs.AddFile("seed.txt", make([]byte, 50)); err != nil {
		t.Fatal(err)
	}
	if got := free(); got != 850 {
		t.Errorf("FreeBytes after replacing seed.txt with 50 bytes = %d, want 850", got)
	}
	if err := fs.Rename("seed.txt", "b.bin", true); err != nil {
		t.Fatal(err)
	}
	if got := free(); got != 950 {
		t.Errorf("FreeBytes after renaming over b.bin = %d, want 950", got)
	}
	if _, err := fs.Open("b.bin", OpenFlags{Write: true, Truncate: true}); err != nil {
		t.Fatal(err)
	}
	if got := free(); got != 1000 {
		t.Errorf("FreeBytes after truncating on open = %d, want 1000", got)
	}
}

// TestMemoryFileSystemRenameKeepsCapacity asserts a rename that would replace a
// directory is refused: dropping only the directory's own entry would orphan its
// children and leave their bytes counted against the volume for good.
func TestMemoryFileSystemRenameKeepsCapacity(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	fs.volume.TotalBytes = 10000

	for _, dir := range []string{"src", "dst"} {
		if err := fs.Mkdir(dir); err != nil {
			t.Fatal(err)
		}
	}
	if err := fs.AddFile("src/x", make([]byte, 100)); err != nil {
		t.Fatal(err)
	}
	if err := fs.AddFile("dst/x", make([]byte, 4096)); err != nil {
		t.Fatal(err)
	}

	if err := fs.Rename("src", "dst", true); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("renaming a directory over a directory returned %v, want ErrAccessDenied", err)
	}
	if err := fs.Rename("src/x", "dst", true); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("renaming a file over a directory returned %v, want ErrAccessDenied", err)
	}

	for _, path := range []string{"src/x", "dst/x"} {
		if err := fs.Remove(path); err != nil {
			t.Fatalf("Remove(%s) = %v", path, err)
		}
	}
	volume, err := fs.VolumeInfo()
	if err != nil {
		t.Fatal(err)
	}
	if volume.FreeBytes != volume.TotalBytes {
		t.Errorf("FreeBytes = %d after every file was removed, want %d", volume.FreeBytes, volume.TotalBytes)
	}
}

// TestMemoryFileSystemTruncateRefusesDirectory asserts a directory handle cannot
// be given data, which Rmdir would never release.
func TestMemoryFileSystemTruncateRefusesDirectory(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.Mkdir("dir"); err != nil {
		t.Fatal(err)
	}
	dir, err := fs.Open("dir", OpenFlags{Read: true})
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()
	if err := dir.Truncate(64); !errors.Is(err, ErrIsDirectory) {
		t.Fatalf("Truncate on a directory returned %v, want ErrIsDirectory", err)
	}
}
