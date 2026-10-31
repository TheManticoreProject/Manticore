package server_test

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// TestServerLocalFileSystemDanglingSymlink asserts that an SMB2 create cannot
// leave a share backed by the host's file system through a symbolic link whose
// target does not exist: FILE_OPEN_IF on such a link created the target outside
// the share.
func TestServerLocalFileSystemDanglingSymlink(t *testing.T) {
	base := t.TempDir()
	share, outside := filepath.Join(base, "share"), filepath.Join(base, "outside")
	for _, dir := range []string{share, outside} {
		if err := os.Mkdir(dir, 0o750); err != nil {
			t.Fatalf("Mkdir() error = %v", err)
		}
	}
	if err := os.Symlink(filepath.Join(outside, "planted.txt"), filepath.Join(share, "link")); err != nil {
		t.Skipf("symbolic links are unavailable here: %v", err)
	}
	fs, err := smb1.NewLocalFileSystem(share, "FILES")
	if err != nil {
		t.Fatalf("NewLocalFileSystem() error = %v", err)
	}

	srv, err := server.NewServer(withTestIdentity(&server.Config{}))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if err := srv.AddShare(&server.Share{Name: "files", FS: fs}); err != nil {
		t.Fatal(err)
	}
	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(listener)
	t.Cleanup(func() { srv.Close() })

	c, err := dial(t, listener.Addr().(*net.TCPAddr).Port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if id, err := c.CreateFile("link", fileflags.GENERIC_WRITE, 7, fileflags.FILE_OPEN_IF, 0); err == nil {
		c.WriteFile(id, 0, []byte("escaped"))
		c.CloseFile(id)
		t.Error("FILE_OPEN_IF succeeded through a dangling link")
	}
	if _, err := os.Lstat(filepath.Join(outside, "planted.txt")); err == nil {
		t.Fatal("an SMB2 create through a dangling link made a file outside the share")
	}
}
