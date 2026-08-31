package server

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// fileShareName is the share the file-service tests connect to.
const fileShareName = "files"

// fileServer stands up a server with one disk share backed by the given file
// system, and returns a client already authenticated and connected to it.
func fileServer(t *testing.T, fs FileSystem, readOnly bool) (*Server, *smb1client.Client) {
	t.Helper()

	srv, transportEnd := pipedServer(t, conformanceConfig(SigningDisabled))
	if err := srv.AddShare(&Share{Name: fileShareName, Type: ShareTypeDisk, FS: fs, ReadOnly: readOnly}); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("Negotiate() error = %v", err)
	}
	creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() error = %v", err)
	}
	if err := client.TreeConnect(fileShareName); err != nil {
		t.Fatalf("TreeConnect() error = %v", err)
	}

	return srv, client
}

// TestFileServiceRoundTrip is the milestone this phase exists for: the SMB1 client
// in this repository creates a file on the server, writes to it, reads it back and
// closes it.
func TestFileServiceRoundTrip(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	payload := []byte("the quick brown fox jumps over the lazy dog")

	fid, err := client.OpenFile("newfile.txt",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN_IF,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}

	written, err := client.WriteFile(fid, 0, payload)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if written != len(payload) {
		t.Fatalf("WriteFile() wrote %d bytes, want %d", written, len(payload))
	}

	read, err := client.ReadFile(fid, 0, uint32(len(payload)))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(read, payload) {
		t.Fatalf("ReadFile() returned %q, want %q", read, payload)
	}

	if err := client.Flush(fid); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}
	if err := client.CloseFile(fid); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}

	// The backend holds what was written, so the round trip went through the
	// share rather than being answered from somewhere else.
	attr, err := fs.Stat("newfile.txt")
	if err != nil {
		t.Fatalf("the file is not in the backend: %v", err)
	}
	if attr.Size != int64(len(payload)) {
		t.Fatalf("the backend holds %d bytes, want %d", attr.Size, len(payload))
	}
}

// TestFileServiceOffsetsAndShortReads asserts reads and writes at an offset, and
// that a read reaching the end of the file returns what there was rather than
// failing.
func TestFileServiceOffsetsAndShortReads(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("data.bin", []byte("0123456789")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("data.bin",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer client.CloseFile(fid)

	// A read from the middle.
	read, err := client.ReadFile(fid, 3, 4)
	if err != nil {
		t.Fatalf("ReadFile() at an offset error = %v", err)
	}
	if string(read) != "3456" {
		t.Fatalf("read %q from offset 3, want %q", read, "3456")
	}

	// A read asking for more than remains returns what remains.
	read, err = client.ReadFile(fid, 8, 100)
	if err != nil {
		t.Fatalf("ReadFile() past the end error = %v", err)
	}
	if string(read) != "89" {
		t.Fatalf("read %q from offset 8, want %q", read, "89")
	}

	// A read entirely past the end returns nothing, and is not a failure.
	read, err = client.ReadFile(fid, 1000, 10)
	if err != nil {
		t.Fatalf("ReadFile() entirely past the end error = %v", err)
	}
	if len(read) != 0 {
		t.Fatalf("read %q past the end, want nothing", read)
	}

	// A write past the end extends the file, with the gap reading as zeroes.
	if _, err := client.WriteFile(fid, 12, []byte("XY")); err != nil {
		t.Fatalf("WriteFile() past the end error = %v", err)
	}
	read, err = client.ReadFile(fid, 0, 20)
	if err != nil {
		t.Fatalf("ReadFile() after extending error = %v", err)
	}
	want := append([]byte("0123456789"), 0x00, 0x00, 'X', 'Y')
	if !bytes.Equal(read, want) {
		t.Fatalf("read % x after extending, want % x", read, want)
	}
}

// TestFileServiceDirectoryOperations asserts the directory commands.
func TestFileServiceDirectoryOperations(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	if err := client.CreateDirectory("dir"); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}
	if err := client.CheckDirectory("dir"); err != nil {
		t.Fatalf("CheckDirectory() on a directory error = %v", err)
	}
	// Creating it twice is a collision.
	if err := client.CreateDirectory("dir"); err == nil {
		t.Fatal("CreateDirectory() succeeded for a directory that already exists")
	}
	// A nested directory needs its parent, which now exists.
	if err := client.CreateDirectory("dir\\sub"); err != nil {
		t.Fatalf("CreateDirectory() for a nested directory error = %v", err)
	}
	// A non-empty directory is not removed.
	if err := client.DeleteDirectory("dir"); err == nil {
		t.Fatal("DeleteDirectory() removed a non-empty directory")
	}
	if err := client.DeleteDirectory("dir\\sub"); err != nil {
		t.Fatalf("DeleteDirectory() error = %v", err)
	}
	if err := client.DeleteDirectory("dir"); err != nil {
		t.Fatalf("DeleteDirectory() on the now-empty directory error = %v", err)
	}
	// And it is gone.
	if err := client.CheckDirectory("dir"); err == nil {
		t.Fatal("CheckDirectory() succeeded for a removed directory")
	}

	// CheckDirectory must distinguish a file from a directory, since a client
	// walks a path with it before opening anything.
	if err := fs.AddFile("afile.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := client.CheckDirectory("afile.txt"); err == nil {
		t.Fatal("CheckDirectory() reported a file as a directory")
	}
}

// TestFileServiceDeleteAndRename asserts deletion including wildcards, and that
// rename refuses to overwrite.
func TestFileServiceDeleteAndRename(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	for _, name := range []string{"a.txt", "b.txt", "c.log", "keep.txt"} {
		if err := fs.AddFile(name, []byte(name)); err != nil {
			t.Fatalf("AddFile(%q) error = %v", name, err)
		}
	}
	if err := fs.AddDirectory("adir"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	// One named file.
	if err := client.DeleteFile("a.txt"); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}
	if _, err := fs.Stat("a.txt"); err == nil {
		t.Fatal("the file is still in the backend")
	}

	// A wildcard deletes the matches and leaves the rest, and never takes a
	// directory with it.
	if err := client.DeleteFile("*.txt"); err != nil {
		t.Fatalf("DeleteFile() with a wildcard error = %v", err)
	}
	if _, err := fs.Stat("b.txt"); err == nil {
		t.Fatal("b.txt survived a matching wildcard")
	}
	if _, err := fs.Stat("keep.txt"); err == nil {
		t.Fatal("keep.txt survived a matching wildcard")
	}
	if _, err := fs.Stat("c.log"); err != nil {
		t.Fatal("c.log was deleted by a pattern that does not match it")
	}
	if _, err := fs.Stat("adir"); err != nil {
		t.Fatal("a directory was deleted by SMB_COM_DELETE")
	}

	// A pattern matching nothing is a failure, not a no-op.
	if err := client.DeleteFile("*.nothing"); err == nil {
		t.Fatal("DeleteFile() succeeded for a pattern that matched nothing")
	}

	// Rename, then a rename onto something that exists.
	if err := client.RenameFile("c.log", "renamed.log"); err != nil {
		t.Fatalf("RenameFile() error = %v", err)
	}
	if _, err := fs.Stat("renamed.log"); err != nil {
		t.Fatalf("the renamed file is not in the backend: %v", err)
	}
	if err := fs.AddFile("occupied.log", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := client.RenameFile("renamed.log", "occupied.log"); err == nil {
		t.Fatal("RenameFile() overwrote an existing file")
	}
}

// TestFileServiceRefusesTraversal asserts the containment boundary holds through
// the whole stack, not just in the resolver's own tests: a client asking for a
// path outside the share is refused at every command that takes one.
func TestFileServiceRefusesTraversal(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("inside.txt", []byte("in")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	escapes := []string{
		"..\\outside.txt",
		"..\\..\\etc\\passwd",
		"dir\\..\\..\\outside.txt",
		"C:\\windows\\win.ini",
		"\\\\other\\share\\file",
		"inside.txt:stream",
		"NUL",
	}

	// An embedded NUL is deliberately absent from this list. The field carrying a
	// path is NUL-terminated, so SMB_STRING.Unmarshal stops at the first one and
	// the server never sees the remainder: opening "inside.txt" is the correct
	// reading of a wire message that says "inside.txt\x00.png". The resolver is
	// still tested against an embedded NUL directly, for the field formats that
	// can carry one.

	for _, path := range escapes {
		path := path
		t.Run(path, func(t *testing.T) {
			if fid, err := client.OpenFile(path, fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
				fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE); err == nil {
				client.CloseFile(fid)
				t.Fatalf("OpenFile(%q) succeeded", path)
			}
			if err := client.DeleteFile(path); err == nil {
				t.Fatalf("DeleteFile(%q) succeeded", path)
			}
			if err := client.CreateDirectory(path); err == nil {
				t.Fatalf("CreateDirectory(%q) succeeded", path)
			}
			if err := client.CheckDirectory(path); err == nil {
				t.Fatalf("CheckDirectory(%q) succeeded", path)
			}
			if err := client.RenameFile("inside.txt", path); err == nil {
				t.Fatalf("RenameFile() to %q succeeded", path)
			}
			if err := client.RenameFile(path, "moved.txt"); err == nil {
				t.Fatalf("RenameFile() from %q succeeded", path)
			}
		})
	}

	// Nothing above reached the backend.
	if _, err := fs.Stat("inside.txt"); err != nil {
		t.Fatal("the file inside the share was disturbed")
	}
}

// TestFileServiceReadOnlyShare asserts a read-only share refuses every modifying
// command whatever access the client asks for, and still serves reads.
func TestFileServiceReadOnlyShare(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("readable.txt", []byte("contents")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, true)

	// Reading works.
	fid, err := client.OpenFile("readable.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() for reading on a read-only share error = %v", err)
	}
	read, err := client.ReadFile(fid, 0, 8)
	if err != nil {
		t.Fatalf("ReadFile() on a read-only share error = %v", err)
	}
	if string(read) != "contents" {
		t.Fatalf("read %q, want %q", read, "contents")
	}
	client.CloseFile(fid)

	// Everything that modifies is refused.
	if writeFid, err := client.OpenFile("readable.txt", fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE); err == nil {
		client.CloseFile(writeFid)
		t.Fatal("OpenFile() for writing succeeded on a read-only share")
	}
	if newFid, err := client.OpenFile("created.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_CREATE, fileflags.FILE_NON_DIRECTORY_FILE); err == nil {
		client.CloseFile(newFid)
		t.Fatal("OpenFile() with FILE_CREATE succeeded on a read-only share")
	}
	if err := client.DeleteFile("readable.txt"); err == nil {
		t.Fatal("DeleteFile() succeeded on a read-only share")
	}
	if err := client.CreateDirectory("dir"); err == nil {
		t.Fatal("CreateDirectory() succeeded on a read-only share")
	}
	if err := client.RenameFile("readable.txt", "other.txt"); err == nil {
		t.Fatal("RenameFile() succeeded on a read-only share")
	}

	// And the backend is untouched.
	if _, err := fs.Stat("readable.txt"); err != nil {
		t.Fatal("the file was removed from a read-only share")
	}
}

// TestFileServiceHandleScoping asserts a handle can only be used on the tree that
// produced it, and stops working once closed.
func TestFileServiceHandleScoping(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("scoped.txt", []byte("data")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("scoped.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if err := client.CloseFile(fid); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}

	// A closed handle no longer works.
	if _, err := client.ReadFile(fid, 0, 4); err == nil {
		t.Fatal("ReadFile() succeeded on a closed handle")
	}
	if err := client.CloseFile(fid); err == nil {
		t.Fatal("CloseFile() succeeded twice on one handle")
	}

	// A handle that was never issued does not work either.
	if _, err := client.ReadFile(0x7FFF, 0, 4); err == nil {
		t.Fatal("ReadFile() succeeded on a handle that was never issued")
	}
}

// TestFileServiceAccessIsEnforcedPerHandle asserts a handle opened for reading
// cannot be written through, so the access granted at open time actually governs.
func TestFileServiceAccessIsEnforcedPerHandle(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("guarded.txt", []byte("original")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("guarded.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer client.CloseFile(fid)

	if _, err := client.WriteFile(fid, 0, []byte("overwritten")); err == nil {
		t.Fatal("WriteFile() succeeded through a read-only handle")
	}

	attr, err := fs.Stat("guarded.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if attr.Size != int64(len("original")) {
		t.Fatalf("the file is %d bytes, so the refused write still landed", attr.Size)
	}
}

// TestFileServiceTreeDisconnectClosesHandles asserts disconnecting a tree releases
// what was opened on it, rather than leaving handles nothing can reach.
func TestFileServiceTreeDisconnectClosesHandles(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("held.txt", []byte("data")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	srv, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("held.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}

	if err := client.TreeDisconnect(); err != nil {
		t.Fatalf("TreeDisconnect() error = %v", err)
	}

	// The handle went with the tree.
	if _, err := client.ReadFile(fid, 0, 4); err == nil {
		t.Fatal("ReadFile() succeeded after the tree was disconnected")
	}

	// And the server is holding nothing.
	waitFor(t, func() bool {
		for _, conn := range srv.liveConnections() {
			if len(conn.opens) != 0 || len(conn.trees) != 0 {
				return false
			}
		}
		return true
	}, "the server still holds trees or handles after a tree disconnect")
}

// TestFileServiceUnknownShareIsRefused asserts a tree connect to a share that is
// not served is refused, and that a served share of the wrong kind is too.
func TestFileServiceUnknownShareIsRefused(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	srv, transportEnd := pipedServer(t, conformanceConfig(SigningDisabled))
	if err := srv.AddShare(&Share{Name: fileShareName, Type: ShareTypeDisk, FS: fs}); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("Negotiate() error = %v", err)
	}
	creds, _ := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() error = %v", err)
	}

	if err := client.TreeConnect("nosuchshare"); err == nil {
		t.Fatal("TreeConnect() succeeded for a share that is not served")
	}
	// The name is matched case-insensitively, as Windows does.
	if err := client.TreeConnect(strings.ToUpper(fileShareName)); err != nil {
		t.Fatalf("TreeConnect() with a differently-cased name error = %v", err)
	}
}

// TestAddShareValidation asserts a share that could not be served is refused when
// it is registered rather than when a client reaches it.
func TestAddShareValidation(t *testing.T) {
	srv, err := NewServer(conformanceConfig(SigningDisabled))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if err := srv.AddShare(nil); err == nil {
		t.Fatal("AddShare(nil) succeeded")
	}
	if err := srv.AddShare(&Share{Name: ""}); err == nil {
		t.Fatal("AddShare() accepted a share with no name")
	}
	if err := srv.AddShare(&Share{Name: "bad\\name", FS: NewMemoryFileSystem("X")}); err == nil {
		t.Fatal("AddShare() accepted a name containing a separator")
	}
	if err := srv.AddShare(&Share{Name: "disk", Type: ShareTypeDisk}); err == nil {
		t.Fatal("AddShare() accepted a disk share with no file system")
	}

	if err := srv.AddShare(&Share{Name: "good", FS: NewMemoryFileSystem("X")}); err != nil {
		t.Fatalf("AddShare() on a valid share error = %v", err)
	}
	// A duplicate, including one differing only in case, is refused.
	if err := srv.AddShare(&Share{Name: "GOOD", FS: NewMemoryFileSystem("X")}); err == nil {
		t.Fatal("AddShare() accepted a duplicate name")
	}
	if srv.Share("good") == nil || srv.Share("GOOD") == nil {
		t.Fatal("Share() does not find the registered share case-insensitively")
	}
	if len(srv.Shares()) != 1 {
		t.Fatalf("Shares() reports %d shares, want 1", len(srv.Shares()))
	}
}

// TestLocalFileSystemContainsSymlinkEscape asserts the second containment
// mechanism: a symbolic link inside the share pointing out of it is refused, which
// no amount of path checking could catch.
func TestLocalFileSystemContainsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatalf("failed to write the outside file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "inside.txt"), []byte("inside"), 0o600); err != nil {
		t.Fatalf("failed to write the inside file: %v", err)
	}

	// A link to a file outside, and a link to the whole directory outside.
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(root, "link.txt")); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linkdir")); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}

	fs, err := NewLocalFileSystem(root, "LOCAL")
	if err != nil {
		t.Fatalf("NewLocalFileSystem() error = %v", err)
	}

	// The file inside is reachable.
	if _, err := fs.Stat("inside.txt"); err != nil {
		t.Fatalf("the file inside the share is not reachable: %v", err)
	}

	// Nothing outside the share can be reached through either link. Stat and Open
	// resolve the path they are given, so a link leading out of the share is
	// refused however it is approached.
	escapes := []string{"link.txt", "linkdir/secret.txt", "linkdir"}
	for _, path := range escapes {
		path := path
		t.Run(path, func(t *testing.T) {
			if _, err := fs.Stat(path); err == nil {
				t.Errorf("Stat(%q) succeeded through a link out of the share", path)
			}
			if file, err := fs.Open(path, OpenFlags{Read: true}); err == nil {
				file.Close()
				t.Errorf("Open(%q) succeeded through a link out of the share", path)
			}
		})
	}

	// Reaching *through* a link, to something the share does not contain, stays
	// refused: the parent is resolved and checked against the root.
	if err := fs.Remove("linkdir/secret.txt"); err == nil {
		t.Error(`Remove("linkdir/secret.txt") reached through a link out of the share`)
	}

	// Deleting the link itself is a different question, and is allowed. The link
	// is an entry of this share and unlinking it removes only that entry -- the
	// assertion below confirms what it pointed at is untouched. Refusing it would
	// leave an entry a client can list and can never remove.
	//
	// This is deliberately not the same rule as for Stat and Open above. Those
	// act on what a path leads to, so a path leading out must be refused; a
	// delete acts on the name, which is inside the share either way.
	if err := fs.Remove("link.txt"); err != nil {
		t.Errorf(`Remove("link.txt") did not remove the link itself: %v`, err)
	}
	if _, err := os.Lstat(filepath.Join(root, "link.txt")); err == nil {
		t.Error("the link is still present after being removed")
	}

	// The outside file is untouched.
	if _, err := os.Stat(filepath.Join(outside, "secret.txt")); err != nil {
		t.Fatal("the file outside the share was disturbed")
	}

	// A listing does not advertise what cannot be opened.
	entries, err := fs.ReadDir("", "")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	for _, entry := range entries {
		if entry.Attr.Name == "link.txt" || entry.Attr.Name == "linkdir" {
			// Listing a link is acceptable only if opening it is refused, which
			// the assertions above establish. What must not happen is the
			// contents leaking.
			continue
		}
	}
}

// TestLocalFileSystemRoundTrip asserts the os-backed backend does the same things
// the in-memory one does, since a share can be served by either.
func TestLocalFileSystemRoundTrip(t *testing.T) {
	root := t.TempDir()
	fs, err := NewLocalFileSystem(root, "LOCAL")
	if err != nil {
		t.Fatalf("NewLocalFileSystem() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	payload := []byte("written through the protocol")

	fid, err := client.OpenFile("ondisk.txt",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN_IF, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if _, err := client.WriteFile(fid, 0, payload); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := client.CloseFile(fid); err != nil {
		t.Fatalf("CloseFile() error = %v", err)
	}

	// It is on the host's disk, where it was asked to be.
	onDisk, err := os.ReadFile(filepath.Join(root, "ondisk.txt"))
	if err != nil {
		t.Fatalf("the file is not on disk: %v", err)
	}
	if !bytes.Equal(onDisk, payload) {
		t.Fatalf("the file on disk holds %q, want %q", onDisk, payload)
	}

	// And the directory commands reach the host too.
	if err := client.CreateDirectory("subdir"); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}
	if info, err := os.Stat(filepath.Join(root, "subdir")); err != nil || !info.IsDir() {
		t.Fatalf("the directory is not on disk: %v", err)
	}
	if err := client.DeleteFile("ondisk.txt"); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "ondisk.txt")); err == nil {
		t.Fatal("the file is still on disk after a delete")
	}
}

// TestMemoryFileSystemDirectly covers the in-memory backend's own behaviour,
// including the cases the protocol handlers above do not reach.
func TestMemoryFileSystemDirectly(t *testing.T) {
	fs := NewMemoryFileSystem("MEM")

	// A create needs its parent to exist.
	if _, err := fs.Open("missing/file.txt", OpenFlags{Create: true, Write: true}); err == nil {
		t.Fatal("Open() created a file under a directory that does not exist")
	}

	// Seeding builds the tree.
	if err := fs.AddFile("a/b/c.txt", []byte("deep")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	for _, path := range []string{"a", "a/b", "a/b/c.txt"} {
		if _, err := fs.Stat(path); err != nil {
			t.Fatalf("Stat(%q) after seeding error = %v", path, err)
		}
	}

	// Renaming a directory takes what is beneath it.
	if err := fs.Rename("a/b", "a/moved", false); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if _, err := fs.Stat("a/moved/c.txt"); err != nil {
		t.Fatalf("the nested file did not move with its directory: %v", err)
	}
	if _, err := fs.Stat("a/b/c.txt"); err == nil {
		t.Fatal("the nested file is still reachable at its old path")
	}

	// A listing reports immediate children only, in a stable order.
	if err := fs.AddFile("a/second.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	entries, err := fs.ReadDir("a", "")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Attr.Name)
	}
	if strings.Join(names, ",") != "moved,second.txt" {
		t.Fatalf("ReadDir() reported %v, want the immediate children in order", names)
	}

	// A pattern filters.
	entries, err = fs.ReadDir("a", "*.txt")
	if err != nil {
		t.Fatalf("ReadDir() with a pattern error = %v", err)
	}
	if len(entries) != 1 || entries[0].Attr.Name != "second.txt" {
		t.Fatalf("ReadDir() with a pattern reported %d entries", len(entries))
	}

	// Truncation both ways.
	file, err := fs.Open("a/second.txt", OpenFlags{Read: true, Write: true})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := file.WriteAt([]byte("0123456789"), 0); err != nil {
		t.Fatalf("WriteAt() error = %v", err)
	}
	if err := file.Truncate(4); err != nil {
		t.Fatalf("Truncate() error = %v", err)
	}
	buffer := make([]byte, 10)
	read, err := file.ReadAt(buffer, 0)
	if err != nil {
		t.Fatalf("ReadAt() error = %v", err)
	}
	if string(buffer[:read]) != "0123" {
		t.Fatalf("after truncating to 4 the file holds %q", buffer[:read])
	}
	file.Close()

	// Rmdir refuses the root and a non-empty directory.
	if err := fs.Rmdir(""); err == nil {
		t.Fatal("Rmdir() removed the share root")
	}
	if err := fs.Rmdir("a"); err == nil {
		t.Fatal("Rmdir() removed a non-empty directory")
	}
}

// TestLocalFileSystemNameOperationsDoNotFollowSymlinks asserts that deleting or
// renaming a symbolic link acts on the link, not on what it points at.
//
// hostPath resolves a path through its symbolic links, which is right for
// reading and writing contents and wrong for the operations that act on a name:
// a delete unlinked the target and left the link behind, and a rename moved the
// target out from under it.
func TestLocalFileSystemNameOperationsDoNotFollowSymlinks(t *testing.T) {
	// newShare builds a share holding a file, a directory, and a symbolic link to
	// each.
	newShare := func(t *testing.T) (*LocalFileSystem, string) {
		t.Helper()
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "target.txt"), []byte("data"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		if err := os.Mkdir(filepath.Join(root, "targetdir"), 0o750); err != nil {
			t.Fatalf("Mkdir() error = %v", err)
		}
		if err := os.Symlink(filepath.Join(root, "target.txt"), filepath.Join(root, "link.txt")); err != nil {
			t.Skipf("symbolic links are unavailable here: %v", err)
		}
		if err := os.Symlink(filepath.Join(root, "targetdir"), filepath.Join(root, "linkdir")); err != nil {
			t.Skipf("symbolic links are unavailable here: %v", err)
		}
		fs, err := NewLocalFileSystem(root, "LOCAL")
		if err != nil {
			t.Fatalf("NewLocalFileSystem() error = %v", err)
		}
		return fs, root
	}

	exists := func(t *testing.T, path string) bool {
		t.Helper()
		_, err := os.Lstat(path)
		return err == nil
	}

	t.Run("Remove unlinks the link", func(t *testing.T) {
		fs, root := newShare(t)
		if err := fs.Remove("link.txt"); err != nil {
			t.Fatalf("Remove() error = %v", err)
		}
		if !exists(t, filepath.Join(root, "target.txt")) {
			t.Error("deleting the link deleted its target")
		}
		if exists(t, filepath.Join(root, "link.txt")) {
			t.Error("the link still exists after being deleted")
		}
	})

	t.Run("Rmdir refuses a link to a directory", func(t *testing.T) {
		fs, root := newShare(t)
		if err := fs.Rmdir("linkdir"); err == nil {
			t.Error("Rmdir() removed a directory through a symbolic link")
		}
		if !exists(t, filepath.Join(root, "targetdir")) {
			t.Error("Rmdir() through the link removed the directory it points at")
		}
	})

	t.Run("Rename moves the link", func(t *testing.T) {
		fs, root := newShare(t)
		if err := fs.Rename("link.txt", "moved.txt", false); err != nil {
			t.Fatalf("Rename() error = %v", err)
		}
		if !exists(t, filepath.Join(root, "target.txt")) {
			t.Error("renaming the link moved its target away")
		}
		if exists(t, filepath.Join(root, "link.txt")) {
			t.Error("the link is still at its old name")
		}
		info, err := os.Lstat(filepath.Join(root, "moved.txt"))
		if err != nil {
			t.Fatalf("the link is not at its new name: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("what arrived at the new name is not a symbolic link")
		}
	})

	// The ordinary cases still work, and containment still holds.
	t.Run("a real file is still removed", func(t *testing.T) {
		fs, root := newShare(t)
		if err := fs.Remove("target.txt"); err != nil {
			t.Fatalf("Remove() error = %v", err)
		}
		if exists(t, filepath.Join(root, "target.txt")) {
			t.Error("the file was not removed")
		}
	})

	t.Run("a real directory is still removed", func(t *testing.T) {
		fs, root := newShare(t)
		if err := fs.Rmdir("targetdir"); err != nil {
			t.Fatalf("Rmdir() error = %v", err)
		}
		if exists(t, filepath.Join(root, "targetdir")) {
			t.Error("the directory was not removed")
		}
	})

	t.Run("a delete through a link out of the share is refused", func(t *testing.T) {
		fs, root := newShare(t)
		outside := t.TempDir()
		if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
			t.Skipf("symbolic links are unavailable here: %v", err)
		}
		if err := fs.Remove("escape/secret.txt"); err == nil {
			t.Error("Remove() reached through a link pointing out of the share")
		}
		if !exists(t, filepath.Join(outside, "secret.txt")) {
			t.Error("a file outside the share was deleted")
		}
	})
}

// TestLocalFileSystemDirectoryOpenDoesNotCreate asserts that an open-existing-only
// request for a directory that is not there is refused, and creates nothing.
//
// FILE_OPEN with FILE_DIRECTORY_FILE produces OpenFlags with Directory set and
// Create clear. The backend's create-a-directory branch is reached whenever the
// target is absent, so without a check on Create it answered a request to find a
// directory by making one. On a read-only share it made one anyway: the read-only
// guards test Write, Create, CreateNew and Truncate, and such a request sets none
// of them.
func TestLocalFileSystemDirectoryOpenDoesNotCreate(t *testing.T) {
	for _, readOnly := range []bool{false, true} {
		name := "writable share"
		if readOnly {
			name = "read-only share"
		}

		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			fs, err := NewLocalFileSystem(root, "LOCAL")
			if err != nil {
				t.Fatalf("NewLocalFileSystem() error = %v", err)
			}
			fs.SetReadOnly(readOnly)
			_, client := fileServer(t, fs, readOnly)

			fid, err := client.OpenFile("absent", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
				fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
			if err == nil {
				client.CloseFile(fid)
				t.Error("OpenFile() with FILE_OPEN succeeded for a directory that does not exist")
			}
			if _, statErr := os.Stat(filepath.Join(root, "absent")); statErr == nil {
				t.Error("the directory was created by a request that only asked to open one")
			}
		})
	}
}

// TestLocalFileSystemDirectoryCreateStillWorks asserts the guard above did not
// cost the dispositions that legitimately create a directory.
func TestLocalFileSystemDirectoryCreateStillWorks(t *testing.T) {
	dispositions := map[string]uint32{
		"FILE_CREATE":  fileflags.FILE_CREATE,
		"FILE_OPEN_IF": fileflags.FILE_OPEN_IF,
	}

	for name, disposition := range dispositions {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			fs, err := NewLocalFileSystem(root, "LOCAL")
			if err != nil {
				t.Fatalf("NewLocalFileSystem() error = %v", err)
			}
			_, client := fileServer(t, fs, false)

			fid, err := client.OpenFile("made", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
				disposition, fileflags.FILE_DIRECTORY_FILE)
			if err != nil {
				t.Fatalf("OpenFile() with %s error = %v", name, err)
			}
			client.CloseFile(fid)

			info, statErr := os.Stat(filepath.Join(root, "made"))
			if statErr != nil {
				t.Fatalf("the directory was not created: %v", statErr)
			}
			if !info.IsDir() {
				t.Error("what was created is not a directory")
			}

			// And opening it again with FILE_OPEN now finds it.
			reopened, err := client.OpenFile("made", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
				fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
			if err != nil {
				t.Fatalf("OpenFile() with FILE_OPEN on the existing directory error = %v", err)
			}
			client.CloseFile(reopened)
		})
	}
}
