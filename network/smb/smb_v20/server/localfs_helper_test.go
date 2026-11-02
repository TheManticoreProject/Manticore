package server_test

import (
	"testing"

	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
)

// newLocalFS builds a LocalFileSystem that is closed when the test ends. The
// file system holds a handle on its root, and Windows refuses to delete a
// directory with an open handle, so the test's temporary directory could not be
// removed otherwise.
func newLocalFS(t testing.TB, root, label string) (*smb1.LocalFileSystem, error) {
	t.Helper()
	fs, err := smb1.NewLocalFileSystem(root, label)
	if err == nil {
		t.Cleanup(func() { fs.Close() })
	}
	return fs, err
}
