package server

import (
	"errors"
	"io"
	"math"
	"path"
	"strings"
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

// TestMemoryFileSystemRenameOntoItself asserts renaming a file onto its own name
// with replace keeps the file instead of deleting it.
func TestMemoryFileSystemRenameOntoItself(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("same.txt", []byte("keep me")); err != nil {
		t.Fatal(err)
	}
	before, err := fs.VolumeInfo()
	if err != nil {
		t.Fatal(err)
	}
	for _, replace := range []bool{true, false} {
		if err := fs.Rename("same.txt", "same.txt", replace); err != nil {
			t.Fatalf("Rename onto itself (replace=%v) = %v, want nil", replace, err)
		}
	}
	file, err := fs.Open("same.txt", OpenFlags{Read: true})
	if err != nil {
		t.Fatalf("the file is gone after a rename onto itself: %v", err)
	}
	defer file.Close()
	contents := make([]byte, 16)
	n, _ := file.ReadAt(contents, 0)
	if got := string(contents[:n]); got != "keep me" {
		t.Errorf("contents after a rename onto itself = %q, want %q", got, "keep me")
	}
	after, err := fs.VolumeInfo()
	if err != nil {
		t.Fatal(err)
	}
	if after.FreeBytes != before.FreeBytes {
		t.Errorf("FreeBytes changed from %d to %d", before.FreeBytes, after.FreeBytes)
	}
}

// TestMemoryFileSystemSeedOverExistingEntry asserts the seeding helpers refuse
// to replace an entry of the other kind: replacing a file with a directory, or a
// directory with a file, would leave data counted against the volume that no
// name can free.
func TestMemoryFileSystemSeedOverExistingEntry(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	fs.volume.TotalBytes = 1000

	if err := fs.AddFile("c", make([]byte, 5)); err != nil {
		t.Fatal(err)
	}
	if err := fs.AddDirectory("c"); !errors.Is(err, ErrExists) {
		t.Errorf("AddDirectory over a file = %v, want ErrExists", err)
	}
	if err := fs.AddFile("d/x", make([]byte, 7)); err != nil {
		t.Fatal(err)
	}
	if err := fs.AddFile("d", make([]byte, 2000)); !errors.Is(err, ErrIsDirectory) {
		t.Errorf("AddFile over a directory = %v, want ErrIsDirectory", err)
	}
	// Seeding an existing directory again is harmless and keeps its children.
	if err := fs.AddDirectory("d"); err != nil {
		t.Errorf("AddDirectory over a directory = %v, want nil", err)
	}
	if attr, err := fs.Stat("d/x"); err != nil || attr.Size != 7 {
		t.Errorf("Stat(d/x) = %+v, %v; want the 7-byte file", attr, err)
	}

	for _, path := range []string{"c", "d/x"} {
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

// TestMemoryFileSystemHandleFollowsRename asserts an open handle keeps working on
// its file after the file is renamed, by itself or with the directory above it,
// and never touches a new file created at the old name.
func TestMemoryFileSystemHandleFollowsRename(t *testing.T) {
	for _, tc := range []struct {
		name, opened, from, to, renamed string
	}{
		{"file", "a.txt", "a.txt", "b.txt", "b.txt"},
		{"directory", "dir/a.txt", "dir", "moved", "moved/a.txt"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewMemoryFileSystem("FILES")
			fs.volume.TotalBytes = 1000
			if err := fs.AddFile(tc.opened, []byte("original")); err != nil {
				t.Fatal(err)
			}
			file, err := fs.Open(tc.opened, OpenFlags{Read: true, Write: true})
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			changes, stop, err := fs.Watch("", true)
			if err != nil {
				t.Fatal(err)
			}
			defer stop()
			if err := fs.Rename(tc.from, tc.to, false); err != nil {
				t.Fatal(err)
			}
			if err := fs.AddFile(tc.opened, []byte("newcomer")); err != nil {
				t.Fatal(err)
			}
			for len(changes) > 0 {
				<-changes
			}

			if _, err := file.WriteAt([]byte("ORIG"), 0); err != nil {
				t.Fatalf("WriteAt through the handle after the rename = %v", err)
			}
			select {
			case change := <-changes:
				if want := strings.ReplaceAll(tc.renamed, "/", `\`); change.Name != want {
					t.Errorf("the write was published as a change to %q, want %q", change.Name, want)
				}
			default:
				t.Error("the write published no change")
			}
			if err := file.Truncate(6); err != nil {
				t.Fatalf("Truncate through the handle after the rename = %v", err)
			}
			contents := make([]byte, 16)
			n, err := file.ReadAt(contents, 0)
			if err != nil && !errors.Is(err, io.EOF) {
				t.Fatalf("ReadAt through the handle after the rename = %v", err)
			}
			if got := string(contents[:n]); got != "ORIGin" {
				t.Errorf("ReadAt through the handle = %q, want %q", got, "ORIGin")
			}
			if attr, err := file.Stat(); err != nil || attr.Size != 6 || attr.Name != path.Base(tc.renamed) {
				t.Errorf("Stat through the handle = %+v, %v; want %s of 6 bytes", attr, err, path.Base(tc.renamed))
			}
			if attr, err := fs.Stat(tc.renamed); err != nil || attr.Size != 6 {
				t.Errorf("Stat(%s) = %+v, %v; want the 6 bytes written through the handle", tc.renamed, attr, err)
			}
			if attr, err := fs.Stat(tc.opened); err != nil || attr.Size != int64(len("newcomer")) {
				t.Errorf("Stat(%s) = %+v, %v; the new file at the old name was changed", tc.opened, attr, err)
			}

			for _, name := range []string{tc.renamed, tc.opened} {
				if err := fs.Remove(name); err != nil {
					t.Fatal(err)
				}
			}
			if volume, _ := fs.VolumeInfo(); volume.FreeBytes != volume.TotalBytes {
				t.Errorf("FreeBytes = %d after every file was removed, want %d", volume.FreeBytes, volume.TotalBytes)
			}
		})
	}
}

// TestMemoryFileSystemHandleOnRemovedEntry asserts a handle whose file lost its
// last name, by a delete or by a rename replacing it, fails rather than reaching
// the file now at its name, and cannot give data to the volume it no longer
// counts against.
func TestMemoryFileSystemHandleOnRemovedEntry(t *testing.T) {
	for _, tc := range []struct {
		name   string
		remove func(fs *MemoryFileSystem) error
	}{
		{"removed", func(fs *MemoryFileSystem) error {
			if err := fs.Remove("a.txt"); err != nil {
				return err
			}
			return fs.AddFile("a.txt", []byte("newcomer"))
		}},
		{"replaced", func(fs *MemoryFileSystem) error {
			if err := fs.AddFile("other.txt", []byte("newcomer")); err != nil {
				return err
			}
			return fs.Rename("other.txt", "a.txt", true)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewMemoryFileSystem("FILES")
			fs.volume.TotalBytes = 1000
			if err := fs.AddFile("a.txt", []byte("original")); err != nil {
				t.Fatal(err)
			}
			file, err := fs.Open("a.txt", OpenFlags{Read: true, Write: true})
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			if err := tc.remove(fs); err != nil {
				t.Fatal(err)
			}

			if _, err := file.WriteAt(make([]byte, 100), 0); !errors.Is(err, ErrNotFound) {
				t.Errorf("WriteAt on a removed file = %v, want ErrNotFound", err)
			}
			if err := file.Truncate(200); !errors.Is(err, ErrNotFound) {
				t.Errorf("Truncate on a removed file = %v, want ErrNotFound", err)
			}
			if _, err := file.ReadAt(make([]byte, 8), 0); !errors.Is(err, ErrNotFound) {
				t.Errorf("ReadAt on a removed file = %v, want ErrNotFound", err)
			}
			if _, err := file.Stat(); !errors.Is(err, ErrNotFound) {
				t.Errorf("Stat on a removed file = %v, want ErrNotFound", err)
			}
			if attr, err := fs.Stat("a.txt"); err != nil || attr.Size != int64(len("newcomer")) {
				t.Errorf("Stat(a.txt) = %+v, %v; the file now at the name was changed", attr, err)
			}

			if err := fs.Remove("a.txt"); err != nil {
				t.Fatal(err)
			}
			if volume, _ := fs.VolumeInfo(); volume.FreeBytes != volume.TotalBytes {
				t.Errorf("FreeBytes = %d after every file was removed, want %d", volume.FreeBytes, volume.TotalBytes)
			}
		})
	}
}

// TestMemoryFileSystemRenameIntoOwnSubtree asserts a directory cannot be moved
// beneath itself. The move would cut the directory and everything under it off
// from the tree, leaving no name to reach it by and its bytes counted against
// the volume for good.
func TestMemoryFileSystemRenameIntoOwnSubtree(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	fs.volume.TotalBytes = 10000

	for _, dir := range []string{"a", "a/b"} {
		if err := fs.Mkdir(dir); err != nil {
			t.Fatal(err)
		}
	}
	if err := fs.AddFile("a/f", []byte("12345")); err != nil {
		t.Fatal(err)
	}

	for _, target := range []string{"a/b/c", "a/c", "a/b/c/d"} {
		if err := fs.Rename("a", target, false); !errors.Is(err, ErrAccessDenied) {
			t.Fatalf("Rename(a, %s) = %v, want ErrAccessDenied", target, err)
		}
		if err := fs.Rename("a", target, true); !errors.Is(err, ErrAccessDenied) {
			t.Fatalf("Rename(a, %s, replace) = %v, want ErrAccessDenied", target, err)
		}
	}
	// The share root holds every other path, so it cannot be moved either.
	if err := fs.Rename("", "a/x", false); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("Rename(root, a/x) = %v, want ErrAccessDenied", err)
	}
	for _, path := range []string{"", "a", "a/b", "a/f"} {
		if _, err := fs.Stat(path); err != nil {
			t.Errorf("Stat(%q) after the refused rename = %v", path, err)
		}
	}

	// A sibling sharing the name as a prefix is not beneath it.
	if err := fs.Rename("a", "ab", false); err != nil {
		t.Fatalf("Rename(a, ab) = %v, want nil", err)
	}
	if err := fs.Remove("ab/f"); err != nil {
		t.Fatalf("Remove(ab/f) = %v", err)
	}
	for _, dir := range []string{"ab/b", "ab"} {
		if err := fs.Rmdir(dir); err != nil {
			t.Fatalf("Rmdir(%s) = %v", dir, err)
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
