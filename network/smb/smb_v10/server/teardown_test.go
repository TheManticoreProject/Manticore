package server

import (
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// waitForConnections waits until the server is serving exactly want
// connections. A connection is untracked only once its receive loop has
// returned, so this also waits for whatever that loop does on the way out.
func waitForConnections(t *testing.T, srv *Server, want int) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for srv.Connections() != want {
		if time.Now().After(deadline) {
			t.Fatalf("the server is serving %d connections, want %d", srv.Connections(), want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestDroppedConnectionReleasesItsHandles asserts that a connection which ends
// without closing its handles still releases what those handles held.
//
// [MS-CIFS] section 3.3.7.2: on a disconnect, "For each Session in
// Connection.SessionTable, the server MUST close the session", and closing a
// session closes every open it holds. A byte-range lock lives in
// a table the whole share consults, so a lock left behind by a client that went
// away would refuse every other client the range until the server restarted.
func TestDroppedConnectionReleasesItsHandles(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("locked.txt", []byte("0123456789abcdef")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddFile("doomed.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	srv, survivor := fileServer(t, fs, false)
	leaver := attachClient(t, srv)

	open := func(c *client.Client, name string, options uint32) client.FID {
		t.Helper()
		fid, err := c.OpenFile(name,
			fileflags.GENERIC_READ|fileflags.GENERIC_WRITE|fileflags.DELETE,
			fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE|fileflags.FILE_SHARE_DELETE,
			fileflags.FILE_OPEN,
			fileflags.FILE_NON_DIRECTORY_FILE|options)
		if err != nil {
			t.Fatalf("opening %s failed: %v", name, err)
		}
		return fid
	}

	held := open(leaver, "locked.txt", 0)
	if err := leaver.LockFile(held, 0, 16, true); err != nil {
		t.Fatalf("locking [0,16) failed: %v", err)
	}
	open(leaver, "doomed.txt", fileflags.FILE_DELETE_ON_CLOSE)

	wanted := open(survivor, "locked.txt", 0)
	if err := survivor.LockFile(wanted, 0, 16, true); err == nil {
		t.Fatal("the other connection locked a range the first one holds")
	}

	// The client goes away without closing anything.
	leaver.Transport.Close()
	waitForConnections(t, srv, 1)

	if err := survivor.LockFile(wanted, 0, 16, true); err != nil {
		t.Errorf("the range was still locked after its holder's connection dropped: %v", err)
	}
	if _, err := fs.Stat("doomed.txt"); err == nil {
		t.Error("a delete-on-close file outlived its handle's connection")
	}
}
