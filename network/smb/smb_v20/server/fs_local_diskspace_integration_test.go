//go:build integration

// Live interoperability check for the LocalFileSystem volume-size report.
// Requires SMB2_TEST_CLIENT (smbclient), like the other integration tests.
package server_test

import (
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
)

// blocksAvailableRe matches smbclient's "<total> blocks of size <n>. <avail>
// blocks available" trailer printed after a directory listing.
var blocksAvailableRe = regexp.MustCompile(`([0-9]+) blocks of size [0-9]+\. ([0-9]+) blocks available`)

// TestLiveSmbclientLocalFSReportsDiskSpace serves a LocalFileSystem share and
// confirms smbclient sees a non-zero disk, guarding the former 0-byte report.
func TestLiveSmbclientLocalFSReportsDiskSpace(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello world"), 0o600); err != nil {
		t.Fatal(err)
	}
	fs, err := smb1.NewLocalFileSystem(root, "FILES")
	if err != nil {
		t.Fatalf("NewLocalFileSystem: %v", err)
	}

	config := withTestIdentity(&server.Config{})
	srv, err := server.NewServer(config)
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
	port := listener.Addr().(*net.TCPAddr).Port

	out := smbclient(t, port, "-c", "ls")

	m := blocksAvailableRe.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("smbclient did not report volume size:\n%s", out)
	}
	total, _ := strconv.ParseInt(m[1], 10, 64)
	avail, _ := strconv.ParseInt(m[2], 10, 64)
	if total <= 0 || avail <= 0 {
		t.Fatalf("volume reported total=%d available=%d blocks, want both > 0:\n%s", total, avail, out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "NT_STATUS_") {
			t.Errorf("smbclient reported an error: %s", line)
		}
	}
	t.Logf("smbclient saw %d total / %d available blocks", total, avail)
}
