package server_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// listNames returns the file names a wildcard QUERY_DIRECTORY reports for a path.
func queryNames(t *testing.T, port int, dir, pattern string) []string {
	t.Helper()
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	id, err := c.CreateFile(dir, fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("open %q: %v", dir, err)
	}
	defer c.CloseFile(id)
	listing, err := c.QueryDirectory(id, uint8(infoclass.FileIdBothDirectoryInformation), pattern, 0)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range filesystem.ParseFileIdBothDirectoryInformationList(listing) {
		names = append(names, e.FileName)
	}
	return names
}

// TestQueryDirectorySynthesisesDotEntries checks the SMB2 server lists "." and
// ".." for a wildcard listing of the share root (Windows does; omitting them was
// a fingerprint and made the SMB1 and SMB2 listings disagree), that the two come
// first, and that a specific-name query does not get them.
func TestQueryDirectorySynthesisesDotEntries(t *testing.T) {
	_, port := startServer(t, server.Config{})

	root := queryNames(t, port, "", "*")
	if len(root) < 2 || root[0] != "." || root[1] != ".." {
		t.Fatalf("root listing = %v, want it to start with \".\", \"..\"", root)
	}
	if !contains(root, "hello.txt") {
		t.Fatalf("root listing %v is missing the real file", root)
	}

	// A specific-name query matches only the file, not the dot entries.
	named := queryNames(t, port, "", "hello.txt")
	if strings.Join(named, ",") != "hello.txt" {
		t.Fatalf("specific-name listing = %v, want just hello.txt", named)
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
