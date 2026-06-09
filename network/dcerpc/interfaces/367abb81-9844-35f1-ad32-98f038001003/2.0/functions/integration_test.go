//go:build integration

// Live integration / wire-validation for svcctl (MS-SCMR). Excluded from the default build
// by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.30 DCERPC_TEST_USER=Administrator DCERPC_TEST_PASS=admin \
//	go test -tags integration -v \
//	  ./network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/functions/
//
// The test opens the SCM (ROpenSCManagerW), opens a well-known service (ROpenServiceW),
// queries its status and config (RQueryServiceStatus / RQueryServiceConfigW), closes both
// handles (RCloseServiceHandle), and probes the enumeration path (REnumServicesStatusW with
// a zero-size buffer to draw ERROR_MORE_DATA + the required byte count). SCM/service context
// handles chain (scm -> service -> close), so the whole sequence runs on ONE association.
//
// Per [MS-SCMR] every method returns ERROR_SUCCESS (0) or a Win32 error ([MS-ERREF]); a clean
// nonzero status still proves the wire is correct, so only a DCE/RPC fault (nca_s_fault_*) is
// treated as a wire failure here.
package functions_test

import (
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/functions"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func ws(s string) ndr.WSTR { return ndr.WSTR(s) }

// isFault reports whether err is a DCE/RPC fault (a wire-modeling failure) rather than a
// clean Win32 error returned by the server (which validates the wire).
func isFault(err error) bool {
	return err != nil && strings.Contains(err.Error(), "fault")
}

func TestIntegration_Svcctl(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live svcctl test")
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

	// SCM and service handles chain, so bind once and run the whole sequence on it.
	rpc := client.NewClient(dcerpcsmb.New(smb, svcctl.PipeName))
	if err := rpc.Bind(svcctl.SyntaxID()); err != nil {
		t.Fatalf("Bind(svcctl): %v", err)
	}
	defer rpc.Close()

	// ROpenSCManagerW (opnum 15): nil machine/database name -> local SCM.
	scm, err := functions.ROpenSCManagerW(rpc, nil, nil,
		svcctl.ScManagerConnect|svcctl.ScManagerEnumerateService)
	if err != nil {
		if isFault(err) {
			t.Fatalf("[WIRE FAIL] ROpenSCManagerW: %v", err)
		}
		t.Fatalf("ROpenSCManagerW: server returned status: %v", err)
	}
	t.Logf("[ok] ROpenSCManagerW: SCM handle acquired")

	// REnumServicesStatusW (opnum 14): a zero-size buffer should return ERROR_MORE_DATA and
	// the required byte count, validating the [out] buffer + [out] count wire shapes.
	_, pcb, _, _, err := functions.REnumServicesStatusW(rpc, scm,
		svcctl.ServiceWin32OwnProcess|svcctl.ServiceWin32ShareProcess, svcctl.ServiceStateAll, 0, nil)
	switch {
	case isFault(err):
		t.Errorf("[WIRE FAIL] REnumServicesStatusW: %v", err)
	case err != nil:
		t.Logf("[ok] REnumServicesStatusW: status (wire validated), needs %d bytes: %v", uint32(pcb), err)
	default:
		t.Logf("[ok] REnumServicesStatusW: succeeded")
	}

	// ROpenServiceW (opnum 16): open a service that exists on every Windows host. The
	// well-known RPC SS service ("RpcSs") is always present.
	const svcName = "RpcSs"
	svc, err := functions.ROpenServiceW(rpc, scm, ws(svcName),
		svcctl.ServiceQueryStatus|svcctl.ServiceQueryConfig)
	if err != nil {
		if isFault(err) {
			t.Fatalf("[WIRE FAIL] ROpenServiceW: %v", err)
		}
		t.Logf("[ok] ROpenServiceW(%s): status (wire validated): %v", svcName, err)
	} else {
		t.Logf("[ok] ROpenServiceW(%s): service handle acquired", svcName)

		// RQueryServiceStatus (opnum 6): read the SERVICE_STATUS record.
		if status, err := functions.RQueryServiceStatus(rpc, svc); err != nil {
			if isFault(err) {
				t.Errorf("[WIRE FAIL] RQueryServiceStatus: %v", err)
			} else {
				t.Logf("[ok] RQueryServiceStatus: status (wire validated): %v", err)
			}
		} else {
			t.Logf("[ok] RQueryServiceStatus: type=0x%x state=%d", uint32(status.DwServiceType), uint32(status.DwCurrentState))
		}

		// RQueryServiceConfigW (opnum 17): zero-size buffer draws ERROR_INSUFFICIENT_BUFFER
		// and the required byte count.
		if _, need, err := functions.RQueryServiceConfigW(rpc, svc, 0); err != nil {
			if isFault(err) {
				t.Errorf("[WIRE FAIL] RQueryServiceConfigW: %v", err)
			} else {
				t.Logf("[ok] RQueryServiceConfigW: status (wire validated), needs %d bytes: %v", uint32(need), err)
			}
		} else {
			t.Logf("[ok] RQueryServiceConfigW: succeeded")
		}

		// RCloseServiceHandle (opnum 0): release the service handle.
		if _, err := functions.RCloseServiceHandle(rpc, svc); err != nil && isFault(err) {
			t.Errorf("[WIRE FAIL] RCloseServiceHandle(service): %v", err)
		} else {
			t.Logf("[ok] RCloseServiceHandle(service)")
		}
	}

	// RCloseServiceHandle (opnum 0): release the SCM handle.
	if _, err := functions.RCloseServiceHandle(rpc, scm); err != nil && isFault(err) {
		t.Errorf("[WIRE FAIL] RCloseServiceHandle(scm): %v", err)
	} else {
		t.Logf("[ok] RCloseServiceHandle(scm)")
	}
}
