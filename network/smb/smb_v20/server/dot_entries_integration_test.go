//go:build integration

// Third-party-peer check that the SMB2 server lists "." and ".." like Windows.
// Requires SMB2_TEST_CLIENT (smbclient), like the other integration tests.
package server_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
)

// TestLiveSmbclientShowsDotEntries confirms a real smbclient sees "." and ".." at
// the share root, guarding the former behaviour where SMB2 listed neither.
func TestLiveSmbclientShowsDotEntries(t *testing.T) {
	_, port := startServer(t, server.Config{})

	out := smbclient(t, port, "-c", "ls")

	for _, want := range []string{" . ", " .. "} {
		if !strings.Contains(out, want) {
			t.Errorf("smbclient listing is missing %q:\n%s", strings.TrimSpace(want), out)
		}
	}
}
