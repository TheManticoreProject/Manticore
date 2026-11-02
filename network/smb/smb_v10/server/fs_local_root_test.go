//go:build !windows

package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// TestLocalFileSystemSwappedDirectoryCannotEscape races a directory inside the
// share being swapped for a symbolic link to a directory outside it against
// creates through that directory. With resolve-then-open, a create could land in
// the outside directory when the swap fell between the two steps; through
// os.Root every component is opened relative to the one before it, so no
// file may ever appear outside.
func TestLocalFileSystemSwappedDirectoryCannotEscape(t *testing.T) {
	base := t.TempDir()
	share := filepath.Join(base, "share")
	outside := filepath.Join(base, "outside")
	for _, d := range []string{share, outside, filepath.Join(share, "d")} {
		if err := os.Mkdir(d, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	fs, err := newLocalFS(t, share, "SHARE")
	if err != nil {
		t.Fatal(err)
	}

	var stop atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		d := filepath.Join(share, "d")
		parked := filepath.Join(share, "parked")
		for !stop.Load() {
			// Swap d for a link out of the share, then back.
			_ = os.Rename(d, parked)
			_ = os.Symlink(outside, d)
			time.Sleep(time.Microsecond)
			_ = os.Remove(d)
			_ = os.Rename(parked, d)
		}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for i := 0; time.Now().Before(deadline); i++ {
		name := fmt.Sprintf("d/f%d", i)
		if f, err := fs.Open(name, OpenFlags{Write: true, Create: true}); err == nil {
			f.Close()
		}
		if _, err := fs.Stat(name); err == nil {
			_ = fs.Remove(name)
		}
		_ = fs.Mkdir(fmt.Sprintf("d/m%d", i))
	}
	stop.Store(true)
	<-done

	escaped, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(escaped) != 0 {
		names := []string{}
		for _, e := range escaped {
			names = append(names, e.Name())
		}
		t.Fatalf("%d entries created outside the share through a swapped directory: %v", len(escaped), names)
	}
}

// TestLocalFileSystemLinkOutOfShareRefused checks a symbolic link inside the share
// pointing outside it is refused for open, stat and listing, and that a create
// through it does not make the target.
func TestLocalFileSystemLinkOutOfShareRefused(t *testing.T) {
	base := t.TempDir()
	share := filepath.Join(base, "share")
	outside := filepath.Join(base, "outside")
	for _, d := range []string{share, outside} {
		if err := os.Mkdir(d, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(share, "out")); err != nil {
		t.Fatal(err)
	}
	fs, err := newLocalFS(t, share, "SHARE")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Open("out/secret", OpenFlags{Read: true}); err == nil {
		t.Error("opened a file through a link out of the share")
	}
	if _, err := fs.Stat("out/secret"); err == nil {
		t.Error("stat through a link out of the share succeeded")
	}
	if _, err := fs.ReadDir("out", "*"); err == nil {
		t.Error("listed a directory through a link out of the share")
	}
	if _, err := fs.Open("out/new", OpenFlags{Write: true, Create: true}); err == nil {
		t.Error("created a file through a link out of the share")
	}
	if _, err := os.Stat(filepath.Join(outside, "new")); err == nil {
		t.Error("a file was created outside the share")
	}
}

// withDeadline fails the test when fn has not returned within two seconds, so a
// call that blocks reports instead of hanging the suite.
func withDeadline(t *testing.T, what string, fn func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		t.Fatalf("%s blocked", what)
		return nil
	}
}

// TestLocalFileSystemFIFORefusedWithoutBlocking checks that a FIFO in the share
// is refused at once by every call that opens it: opening one for reading or
// writing waits for the other end, which would stall the server.
func TestLocalFileSystemFIFORefusedWithoutBlocking(t *testing.T) {
	share := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(share, "pipe"), 0o640); err != nil {
		t.Skipf("mkfifo: %v", err)
	}
	fs, err := newLocalFS(t, share, "SHARE")
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()

	if err := withDeadline(t, "ReadDir", func() error { _, err := fs.ReadDir("pipe", "*"); return err }); !errors.Is(err, ErrNotDirectory) {
		t.Errorf("ReadDir on a FIFO = %v, want ErrNotDirectory", err)
	}
	if err := withDeadline(t, "SetAttr size", func() error { return fs.SetAttr("pipe", FileAttr{Size: 1}, AttrMask{Size: true}) }); !errors.Is(err, ErrAccessDenied) {
		t.Errorf("SetAttr size on a FIFO = %v, want ErrAccessDenied", err)
	}
	if err := withDeadline(t, "SetAttr read-only", func() error {
		return fs.SetAttr("pipe", FileAttr{ReadOnly: true}, AttrMask{ReadOnly: true})
	}); !errors.Is(err, ErrAccessDenied) {
		t.Errorf("SetAttr read-only on a FIFO = %v, want ErrAccessDenied", err)
	}
	for _, flags := range []OpenFlags{{Read: true}, {Write: true}} {
		if err := withDeadline(t, "Open", func() error { _, err := fs.Open("pipe", flags); return err }); !errors.Is(err, ErrAccessDenied) {
			t.Errorf("Open(%+v) on a FIFO = %v, want ErrAccessDenied", flags, err)
		}
	}
}

// TestLocalFileSystemCreateThroughDanglingLinkRefused checks that a create
// through a dangling link inside the share makes nothing: the file created would
// not be the one the client named.
func TestLocalFileSystemCreateThroughDanglingLinkRefused(t *testing.T) {
	share := t.TempDir()
	if err := os.Mkdir(filepath.Join(share, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("sub", "newtarget"), filepath.Join(share, "dangling")); err != nil {
		t.Fatal(err)
	}
	fs, err := newLocalFS(t, share, "SHARE")
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()

	for _, flags := range []OpenFlags{{Write: true, Create: true}, {Write: true, CreateNew: true}} {
		if _, err := fs.Open("dangling", flags); !errors.Is(err, ErrAccessDenied) {
			t.Errorf("Open(%+v) through a dangling link = %v, want ErrAccessDenied", flags, err)
		}
	}
	if _, err := os.Lstat(filepath.Join(share, "sub", "newtarget")); err == nil {
		t.Error("a create through a dangling link made its target")
	}
}

// TestLocalFileSystemAbsoluteLinkInsideShareRefused records a limit of os.Root:
// a link with an absolute target is refused even when the target is inside the
// share. A relative link to the same place still works.
func TestLocalFileSystemAbsoluteLinkInsideShareRefused(t *testing.T) {
	share := t.TempDir()
	if err := os.Mkdir(filepath.Join(share, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(share, "sub", "f"), []byte("x"), 0o640); err != nil {
		t.Fatal(err)
	}
	fs, err := newLocalFS(t, share, "SHARE")
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()
	if err := os.Symlink(filepath.Join(fs.Root(), "sub"), filepath.Join(share, "abslink")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("sub", filepath.Join(share, "rellink")); err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Stat("abslink/f"); !errors.Is(err, ErrAccessDenied) {
		t.Errorf("Stat through an absolute link = %v, want ErrAccessDenied", err)
	}
	if _, err := fs.ReadDir("abslink", "*"); !errors.Is(err, ErrAccessDenied) {
		t.Errorf("ReadDir through an absolute link = %v, want ErrAccessDenied", err)
	}
	if _, err := fs.Stat("rellink/f"); err != nil {
		t.Errorf("Stat through a relative link: %v", err)
	}
	if entries, err := fs.ReadDir("rellink", "*"); err != nil || len(entries) == 0 {
		t.Errorf("ReadDir through a relative link = %d entries, %v", len(entries), err)
	}
}

// TestLocalFileSystemReadOnlyToggleOnUnreadableFile checks that the read-only
// attribute of a write-only file can be set: changing the mode needs a handle,
// not read access.
func TestLocalFileSystemReadOnlyToggleOnUnreadableFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads any file")
	}
	share := t.TempDir()
	path := filepath.Join(share, "wo")
	if err := os.WriteFile(path, []byte("x"), 0o200); err != nil {
		t.Fatal(err)
	}
	fs, err := newLocalFS(t, share, "SHARE")
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()

	if err := fs.SetAttr("wo", FileAttr{ReadOnly: true}, AttrMask{ReadOnly: true}); err != nil {
		t.Fatalf("SetAttr read-only on a write-only file: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o222 != 0 {
		t.Errorf("mode after setting read-only = %v", info.Mode().Perm())
	}
}
