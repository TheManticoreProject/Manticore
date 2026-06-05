//go:build integration

// Live integration / wire-validation for srvsvc over the full DCE/RPC-over-SMB stack.
// Excluded from the default build by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.27 DCERPC_TEST_USER=user DCERPC_TEST_PASS=user \
//	go test -tags integration -v \
//	  ./network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/functions/
package functions_test

import (
	"net"
	"os"
	"strconv"
	"testing"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestIntegration_Srvsvc(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live srvsvc test")
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

	// Each method runs on a FRESH pipe + bind so a fault on one cannot desync the next.
	bind := func(t *testing.T) (*client.Client, func()) {
		t.Helper()
		rpc := client.NewClient(dcerpcsmb.New(smb, srvsvc.PipeName))
		if err := rpc.Bind(srvsvc.SyntaxID()); err != nil {
			t.Fatalf("DCE/RPC Bind(srvsvc): %v", err)
		}
		return rpc, func() { _ = rpc.Close() }
	}

	// NetrRemoteTOD (opnum 28): a pointer to a fixed all-DWORD struct.
	func() {
		rpc, done := bind(t)
		defer done()
		if tod, err := functions.NetrRemoteTOD(rpc, ""); err != nil {
			t.Errorf("[WIRE FAIL] NetrRemoteTOD: %v", err)
		} else if tod != nil {
			t.Logf("[ok] NetrRemoteTOD: %04d-%02d-%02d %02d:%02d:%02d tz=%d",
				tod.TodYear, tod.TodMonth, tod.TodDay, tod.TodHours, tod.TodMins, tod.TodSecs, tod.TodTimezone)
		}
	}()

	// NetrServerGetInfo level 101 (opnum 21): a switch_is union with a single arm.
	func() {
		rpc, done := bind(t)
		defer done()
		if info, err := functions.NetrServerGetInfo(rpc, "", 101); err != nil {
			t.Errorf("[WIRE FAIL] NetrServerGetInfo(101): %v", err)
		} else if info.ServerInfo101 != nil {
			si := info.ServerInfo101
			t.Logf("[ok] NetrServerGetInfo(101): name=%q version=%d.%d type=%#x comment=%q",
				string(si.Sv101Name), si.Sv101VersionMajor, si.Sv101VersionMinor, si.Sv101Type, string(si.Sv101Comment))
		} else {
			t.Errorf("NetrServerGetInfo(101): union arm ServerInfo101 is nil (tag=%d)", info.Tag)
		}
	}()

	// NetrShareEnum level 1 (opnum 15): the heavy enum pattern — an [in,out] ENUM_STRUCT
	// whose union arm is a pointer to a container holding a conformant array of
	// pointer-bearing SHARE_INFO_1 (each with [unique] string members).
	func() {
		rpc, done := bind(t)
		defer done()
		resume := ndr.DWORD(0)
		// The [in,out] union arm must be a non-null (empty) container on input.
		in := structures.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: structures.SHARE_ENUM_UNION{Tag: 1, Level1: &structures.SHARE_INFO_1_CONTAINER{}}}
		out, total, _, err := functions.NetrShareEnum(rpc, "", in, 0xFFFFFFFF, &resume)
		if err != nil {
			t.Errorf("[WIRE FAIL] NetrShareEnum(1): %v", err)
		} else if out.ShareInfo.Level1 != nil {
			shares := out.ShareInfo.Level1.Buffer
			t.Logf("[ok] NetrShareEnum(1): total=%d entriesRead=%d", total, out.ShareInfo.Level1.EntriesRead)
			for i, s := range shares {
				t.Logf("    [%d] %q type=%#x remark=%q", i, string(s.Shi1Netname), s.Shi1Type, string(s.Shi1Remark))
			}
		} else {
			t.Errorf("NetrShareEnum(1): union arm Level1 is nil (tag=%d)", out.ShareInfo.Tag)
		}
	}()
}
