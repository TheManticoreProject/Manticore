//go:build integration

// Live integration / wire-validation for the endpoint mapper (ept) over the full
// DCE/RPC-over-SMB stack (the \epmapper named pipe). Excluded from the default build by
// the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.27 DCERPC_TEST_USER=user DCERPC_TEST_PASS=user \
//	go test -tags integration -v \
//	  ./network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions/
//
// The same resolution also works over ncacn_ip_tcp on port 135 (see
// network/dcerpc/v5/transport/tcp): substitute that transport for the SMB one below.
package functions_test

import (
	"net"
	"os"
	"strconv"
	"testing"

	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

func TestIntegration_EptMap(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live ept_map test")
	}
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
	creds, err := credentials.NewCredentials(os.Getenv("DCERPC_TEST_DOMAIN"), os.Getenv("DCERPC_TEST_USER"), os.Getenv("DCERPC_TEST_PASS"), "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	smb := smbclient.NewClientUsingTCPTransport(ip, port)
	if err := smb.Connect(ip, port); err != nil {
		t.Fatalf("SMB Connect: %v", err)
	}
	defer smb.Disconnect()
	smb.NativeOS, smb.NativeLanMan = "Unix", "Samba"
	if err := smb.SessionSetup(creds); err != nil {
		t.Fatalf("SMB SessionSetup: %v", err)
	}
	defer smb.Logoff()
	if err := smb.TreeConnect("IPC$"); err != nil {
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	defer smb.TreeDisconnect()

	rpc := client.NewClient(dcerpcsmb.New(smb, epm.PipeName))
	if err := rpc.Bind(epm.SyntaxID()); err != nil {
		t.Fatalf("DCE/RPC Bind(ept): %v", err)
	}
	defer rpc.Close()

	// Resolve srvsvc (4b324fc8-1670-01d3-1278-5a47bf6ee188 v3.0); it is registered on
	// every Windows host with a dynamic ncacn_ip_tcp endpoint.
	srvsvc := guid.GUID{A: 0x4b324fc8, B: 0x1670, C: 0x01d3, D: 0x1278, E: 0x5a47bf6ee188}
	eps, err := functions.Map(rpc, srvsvc, 3, 0)
	if err != nil {
		t.Fatalf("[WIRE FAIL] ept_map(srvsvc): %v", err)
	}
	if len(eps) == 0 {
		t.Fatal("[WIRE FAIL] ept_map(srvsvc) returned no ncacn_ip_tcp endpoints")
	}
	for _, ep := range eps {
		t.Logf("[ok] srvsvc bound at %s", ep)
	}
}

func TestIntegration_EptLookup(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live ept_lookup test")
	}
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
	creds, err := credentials.NewCredentials(os.Getenv("DCERPC_TEST_DOMAIN"), os.Getenv("DCERPC_TEST_USER"), os.Getenv("DCERPC_TEST_PASS"), "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	smb := smbclient.NewClientUsingTCPTransport(ip, port)
	if err := smb.Connect(ip, port); err != nil {
		t.Fatalf("SMB Connect: %v", err)
	}
	defer smb.Disconnect()
	smb.NativeOS, smb.NativeLanMan = "Unix", "Samba"
	if err := smb.SessionSetup(creds); err != nil {
		t.Fatalf("SMB SessionSetup: %v", err)
	}
	defer smb.Logoff()
	if err := smb.TreeConnect("IPC$"); err != nil {
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	defer smb.TreeDisconnect()

	rpc := client.NewClient(dcerpcsmb.New(smb, epm.PipeName))
	if err := rpc.Bind(epm.SyntaxID()); err != nil {
		t.Fatalf("DCE/RPC Bind(ept): %v", err)
	}
	defer rpc.Close()

	// Enumerate the whole endpoint map and render each entry as a string binding.
	entries, err := functions.Lookup(rpc)
	if err != nil {
		t.Fatalf("[WIRE FAIL] ept_lookup: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("[WIRE FAIL] ept_lookup returned no entries")
	}
	for _, e := range entries {
		tw, err := e.DecodeTower()
		if err != nil {
			t.Errorf("[WIRE FAIL] decode tower: %v", err)
			continue
		}
		obj := e.Object.GUID()
		// Binding() renders every transport (ncacn_ip_tcp, ncacn_np, ncacn_http, ...); fall
		// back to a floor-count summary for any tower it does not recognize.
		if b, err := tw.Binding(); err == nil {
			t.Logf("[ok] %s  iface=%s  %q", b, obj.ToFormatD(), e.Annotation)
		} else {
			t.Logf("[ok] %d-floor tower  iface=%s  %q", len(tw.Floors), obj.ToFormatD(), e.Annotation)
		}
	}
	t.Logf("[ok] ept_lookup enumerated %d endpoint-map entries", len(entries))
}
