//go:build integration

// Live integration / wire-validation for lsacap (MS-CAPR) over the full
// DCE/RPC-over-SMB stack. Excluded from the default build by the "integration" tag.
// Run with:
//
//	DCERPC_TEST_HOST=192.168.1.27 DCERPC_TEST_USER=user DCERPC_TEST_PASS=user \
//	go test -tags integration -v \
//	  ./network/dcerpc/interfaces/afc07e2e-311c-4435-808c-c483ffeec7c9/1.0/functions/
//
// lsacap defines a single method (LsarGetAvailableCAPIDs, opnum 0) with no context
// handle, so the call is independent and runs on its own bind over the shared
// \lsarpc pipe.
package functions_test

import (
	"net"
	"os"
	"strconv"
	"testing"

	lsacap "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/afc07e2e-311c-4435-808c-c483ffeec7c9/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/afc07e2e-311c-4435-808c-c483ffeec7c9/1.0/functions"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestIntegration_Lsacap(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live lsacap test")
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

	rpc := client.NewClient(dcerpcsmb.New(smb, lsacap.PipeName))
	if err := rpc.Bind(lsacap.SyntaxID()); err != nil {
		t.Fatalf("DCE/RPC Bind(lsacap): %v", err)
	}
	defer func() { _ = rpc.Close() }()

	// LsarGetAvailableCAPIDs (opnum 0): exercises the [out] [ref] LSAPR_WRAPPED_CAPID_SET,
	// a [unique] pointer to a conformant array (Entries) of LSAPR_SID_INFORMATION, each
	// holding a [unique] RPC_SID. A machine with no deployed central access policies
	// returns Entries==0, which still validates the response wire shape end to end.
	set, err := functions.LsarGetAvailableCAPIDs(rpc)
	if err != nil {
		t.Fatalf("[WIRE FAIL] LsarGetAvailableCAPIDs: %v", err)
	}
	t.Logf("[ok] LsarGetAvailableCAPIDs: Entries=%d SidInfo=%d", set.Entries, len(set.SidInfo))
	for i, si := range set.SidInfo {
		if si.Sid != nil {
			t.Logf("    CAPID[%d] = %s", i, si.Sid.String())
		}
	}
}
