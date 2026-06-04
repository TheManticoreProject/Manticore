//go:build integration

// Integration test for srvsvc over the full DCE/RPC-over-SMB stack against a live
// server. Excluded from the default build by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.27 \
//	DCERPC_TEST_USER=user DCERPC_TEST_PASS=user \
//	go test -tags integration -v ./network/dcerpc/interfaces/srvsvc/
//
// Optional: DCERPC_TEST_DOMAIN, DCERPC_TEST_PORT (default 445).
package srvsvc

import (
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func boundClient(t *testing.T) (*client.Client, func()) {
	t.Helper()
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live integration test")
	}
	ip := net.ParseIP(host)
	if ip == nil {
		t.Fatalf("DCERPC_TEST_HOST %q is not a valid IP", host)
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

	smb := smbclient.NewClientUsingTCPTransport(ip, port)
	if err := smb.Connect(ip, port); err != nil {
		t.Fatalf("SMB Connect: %v", err)
	}
	smb.NativeOS = "Unix"
	smb.NativeLanMan = "Samba"
	if err := smb.SessionSetup(creds); err != nil {
		t.Fatalf("SMB SessionSetup: %v", err)
	}
	if err := smb.TreeConnect("IPC$"); err != nil {
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}

	rpc := client.NewClient(dcerpcsmb.New(smb, PipeName))
	if err := rpc.Bind(SyntaxID()); err != nil {
		t.Fatalf("DCE/RPC Bind(srvsvc): %v", err)
	}
	cleanup := func() {
		_ = rpc.Close()
		_ = smb.TreeDisconnect()
		_ = smb.Logoff()
		_ = smb.Disconnect()
	}
	return rpc, cleanup
}

func TestIntegration_RemoteTOD(t *testing.T) {
	rpc, cleanup := boundClient(t)
	defer cleanup()

	tod, err := RemoteTOD(rpc)
	if err != nil {
		t.Fatalf("RemoteTOD: %v", err)
	}
	t.Logf("server time of day: %04d-%02d-%02d %02d:%02d:%02d (weekday %d, tz %d min)",
		tod.Year, tod.Month, tod.Day, tod.Hours, tod.Mins, tod.Secs, tod.Weekday, int32(tod.Timezone))
	if tod.Year < 1990 || tod.Month < 1 || tod.Month > 12 || tod.Hours > 23 {
		t.Errorf("implausible time-of-day values: %+v", tod)
	}
}

func TestIntegration_ServerGetInfo101(t *testing.T) {
	rpc, cleanup := boundClient(t)
	defer cleanup()

	info, err := ServerGetInfo101(rpc)
	if err != nil {
		t.Fatalf("ServerGetInfo101: %v", err)
	}
	name, comment := "", ""
	if info.Name != nil {
		name = string(*info.Name)
	}
	if info.Comment != nil {
		comment = string(*info.Comment)
	}
	t.Logf("server: name=%q version=%d.%d type=0x%08x comment=%q",
		name, info.VersionMajor, info.VersionMinor, info.Type, comment)
	if name == "" {
		t.Error("ServerGetInfo101 returned an empty server name")
	}
}
