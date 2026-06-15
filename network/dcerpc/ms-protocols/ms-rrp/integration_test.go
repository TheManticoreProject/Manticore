//go:build integration

// Live integration tests for the MS-RRP convenience layer over the full DCE/RPC-over-SMB
// stack. Excluded from the default build by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.30 DCERPC_TEST_USER=Administrator DCERPC_TEST_PASS=admin \
//	go test -tags integration -v ./network/dcerpc/ms-protocols/ms-rrp/
//
// The Remote Registry service must be running on the target. The test reads the registry
// version, enumerates a well-known key's subkeys, and reads the ProductName value, all on
// the single bound association RemoteRegistry maintains.
package ms_rrp_test

import (
	"os"
	"strconv"
	"testing"

	ms_rrp "github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/ms-rrp"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func liveRegistry(t *testing.T) (*ms_rrp.RemoteRegistry, func()) {
	t.Helper()
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live MS-RRP test")
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
	if err := smb.Login(creds); err != nil {
		t.Fatalf("SMB Login: %v", err)
	}
	if err := smb.TreeConnect("IPC$"); err != nil {
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	reg := ms_rrp.New(smb)
	if err := reg.Connect(); err != nil {
		t.Fatalf("RemoteRegistry.Connect: %v", err)
	}
	cleanup := func() {
		_ = reg.Close()
		_ = smb.TreeDisconnect()
		_ = smb.Logoff()
		_ = smb.Disconnect()
	}
	return reg, cleanup
}

func TestIntegration_RemoteRegistry(t *testing.T) {
	reg, cleanup := liveRegistry(t)
	defer cleanup()

	hklm, err := reg.OpenLocalMachine(nil, ms_rrp.KeyRead)
	if err != nil {
		t.Fatalf("OpenLocalMachine: %v", err)
	}
	defer reg.BaseRegCloseKey(hklm)

	if ver, err := reg.BaseRegGetVersion(hklm); err != nil {
		t.Errorf("BaseRegGetVersion: %v", err)
	} else {
		t.Logf("[ok] registry version %d", uint32(ver))
	}

	const cv = `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`
	subkeys, err := reg.EnumKeysByPath(cv)
	if err != nil {
		t.Errorf("EnumKeysByPath(%s): %v", cv, err)
	} else {
		t.Logf("[ok] %s has %d subkeys", cv, len(subkeys))
	}

	if v, err := reg.QueryValueByPath(cv, "ProductName"); err != nil {
		t.Errorf("QueryValueByPath ProductName: %v", err)
	} else {
		t.Logf("[ok] ProductName = %q (type %d)", v.String(), v.Type)
	}
}
