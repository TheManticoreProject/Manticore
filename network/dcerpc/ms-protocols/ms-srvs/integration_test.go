//go:build integration

// Live integration tests for the MS-SRVS convenience layer over the full
// DCE/RPC-over-SMB stack. Excluded from the default build by the "integration" tag.
// Run with:
//
//	DCERPC_TEST_HOST=192.168.1.27 DCERPC_TEST_USER=user DCERPC_TEST_PASS=user \
//	go test -tags integration -v ./network/dcerpc/ms-protocols/ms-srvs/
package mssrvs_test

import (
	"os"
	"strconv"
	"testing"

	mssrvs "github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/ms-srvs"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// liveClient builds an MS-SRVS client over a live SMB session using the generic,
// version-agnostic SMB client (so it runs over whichever dialect the server
// negotiates — SMB1 or SMB2). It skips the test unless DCERPC_TEST_HOST is set.
func liveClient(t *testing.T) (*mssrvs.Client, func()) {
	t.Helper()
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live MS-SRVS test")
	}
	port := 445
	if p := os.Getenv("DCERPC_TEST_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			t.Fatalf("invalid DCERPC_TEST_PORT %q: %v", p, err)
		}
		port = n
	}
	creds, err := credentials.NewCredentials(os.Getenv("DCERPC_TEST_DOMAIN"), os.Getenv("DCERPC_TEST_USER"), os.Getenv("DCERPC_TEST_PASS"), "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	smb, err := smbclient.Dial(host, port, smbclient.Options{})
	if err != nil {
		t.Fatalf("SMB Dial: %v", err)
	}
	t.Logf("negotiated %s", smb.Dialect())
	if err := smb.Login(creds); err != nil {
		t.Fatalf("SMB Login: %v", err)
	}
	if err := smb.TreeConnect("IPC$"); err != nil {
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	cleanup := func() {
		_ = smb.TreeDisconnect()
		_ = smb.Logoff()
		_ = smb.Disconnect()
	}
	return mssrvs.New(smb), cleanup
}

func TestIntegration_ListShares(t *testing.T) {
	c, cleanup := liveClient(t)
	defer cleanup()
	shares, err := c.ListShares()
	if err != nil {
		t.Fatalf("[WIRE FAIL] ListShares: %v", err)
	}
	if len(shares) == 0 {
		t.Fatal("[WIRE FAIL] ListShares returned no shares (expected at least IPC$)")
	}
	for _, s := range shares {
		t.Logf("[ok] share %q type=0x%08x %q", s.Name, s.Type, s.Comment)
	}
}
