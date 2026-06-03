//go:build integration

// Integration test for the full DCE/RPC-over-SMB stack against a live server.
//
// It is excluded from the default build by the "integration" tag, since it needs a
// reachable SMB server. Run it with, for example:
//
//	DCERPC_TEST_HOST=192.168.1.27 \
//	DCERPC_TEST_USER=user DCERPC_TEST_PASS=user \
//	go test -tags integration -v ./network/dcerpc/interfaces/lsarpc/
//
// Optional: DCERPC_TEST_DOMAIN, DCERPC_TEST_PORT (default 445).
package lsarpc

import (
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestIntegration_OpenAndClosePolicy(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live integration test")
	}
	user := os.Getenv("DCERPC_TEST_USER")
	pass := os.Getenv("DCERPC_TEST_PASS")
	domain := os.Getenv("DCERPC_TEST_DOMAIN")
	port := 445
	if p := os.Getenv("DCERPC_TEST_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			t.Fatalf("invalid DCERPC_TEST_PORT %q: %v", p, err)
		}
		port = n
	}

	ip := net.ParseIP(host)
	if ip == nil {
		t.Fatalf("DCERPC_TEST_HOST %q is not a valid IP", host)
	}

	creds, err := credentials.NewCredentials(domain, user, pass, "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	// Establish the SMB session and connect to the IPC$ tree that hosts named pipes.
	smb := smbclient.NewClientUsingTCPTransport(ip, port)
	if err := smb.Connect(ip, port); err != nil {
		t.Fatalf("SMB Connect: %v", err)
	}
	defer smb.Disconnect()
	// NativeOS/NativeLanMan must be set before SessionSetup; some servers reject the
	// session setup otherwise (mirrors the repo's main.go).
	smb.NativeOS = "Unix"
	smb.NativeLanMan = "Samba"
	if err := smb.SessionSetup(creds); err != nil {
		t.Fatalf("SMB SessionSetup: %v", err)
	}
	defer smb.Logoff()
	if err := smb.TreeConnect("IPC$"); err != nil {
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	defer smb.TreeDisconnect()

	// Bind DCE/RPC to lsarpc over the named pipe.
	pipe := dcerpcsmb.New(smb, PipeName)
	rpc := client.NewClient(pipe)
	defer rpc.Close()
	if err := rpc.Bind(SyntaxID()); err != nil {
		t.Fatalf("DCE/RPC Bind(lsarpc): %v", err)
	}
	t.Log("bound to lsarpc")

	// LsarOpenPolicy2 -> handle, then LsarClose.
	handle, err := OpenPolicy2(rpc, MaximumAllowed)
	if err != nil {
		t.Fatalf("OpenPolicy2: %v", err)
	}
	t.Logf("OpenPolicy2 returned handle %x", handle)
	if handle.IsZero() {
		t.Error("OpenPolicy2 returned a zero handle on success")
	}

	if _, err := Close(rpc, handle); err != nil {
		t.Fatalf("Close: %v", err)
	}
	t.Log("policy handle closed")
}
