//go:build integration

// Live integration / wire-validation for trkwks (MS-DLTW, the Distributed Link Tracking:
// Workstation Protocol). Excluded from the default build by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.30 DCERPC_TEST_USER=Administrator DCERPC_TEST_PASS=admin \
//	go test -tags integration -v \
//	  ./network/dcerpc/interfaces/300f3532-38cc-11d0-a3f0-0020af6b0add/1.2/functions/
//
// The protocol is named-pipe only (ncacn_np): the test binds over \pipe\trkwks and, if
// that endpoint is not registered, retries over \pipe\ntsvcs — exactly the fallback the
// client is prescribed to perform ([MS-DLTW] 2.1). It then issues LnkSearchMachine with
// zeroed droids; the server will not find such a file, but any clean HRESULT (including
// the TRK_E_* soft failures) proves the request/response wire shape marshals correctly.
// Only a DCE/RPC fault (nca_s_fault_*) is treated as a wire failure.
package functions_test

import (
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	trkwks "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/300f3532-38cc-11d0-a3f0-0020af6b0add/1.2"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/300f3532-38cc-11d0-a3f0-0020af6b0add/1.2/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	msdltw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dltw"
)

// bindTrkwks establishes an SMB session to the target, then binds the trkwks interface
// over a named pipe — preferring \trkwks and falling back to \ntsvcs ([MS-DLTW] 2.1) —
// returning the bound client as an ndr.Invoker.
func bindTrkwks(t *testing.T) ndr.Invoker {
	t.Helper()
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live integration test")
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
	smb.NativeOS, smb.NativeLanMan = "Unix", "Samba"
	if err := smb.SessionSetup(creds); err != nil {
		smb.Disconnect()
		t.Fatalf("SMB SessionSetup: %v", err)
	}
	if err := smb.TreeConnect("IPC$"); err != nil {
		smb.Logoff()
		smb.Disconnect()
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	closeSMB := func() { smb.TreeDisconnect(); smb.Logoff(); smb.Disconnect() }

	for _, pipe := range []string{trkwks.PipeName, trkwks.PipeNameFallback} {
		rpc := client.NewClient(dcerpcsmb.New(smb, pipe))
		if err := rpc.Bind(trkwks.SyntaxID()); err != nil {
			_ = rpc.Close()
			t.Logf("Bind over %s failed: %v", pipe, err)
			continue
		}
		t.Logf("bound to trkwks over %s", pipe)
		t.Cleanup(func() { _ = rpc.Close(); closeSMB() })
		return rpc
	}
	closeSMB()
	t.Skip("neither \\trkwks nor \\ntsvcs exposes the trkwks interface on this host")
	return nil
}

// TestIntegration_LnkSearchMachine issues a search for a nonexistent (all-zero) file. The
// server cannot find it, so a nonzero HRESULT is expected; that clean response — as
// opposed to a DCE/RPC fault — validates that the request and response stubs marshal
// correctly on the wire.
func TestIntegration_LnkSearchMachine(t *testing.T) {
	rpc := bindTrkwks(t)

	var zero msdltw.CDomainRelativeObjId
	birthNext, next, mcidNext, path, err := functions.LnkSearchMachine(rpc, 0, zero, zero)
	t.Logf("LnkSearchMachine: err=%v", err)
	t.Logf("  pdroidBirthNext=%+v", birthNext)
	t.Logf("  pdroidNext=%+v", next)
	t.Logf("  pmcidNext=%q", mcidNext.String())
	t.Logf("  ptszPath=%q", string(path))

	if err != nil && strings.Contains(strings.ToLower(err.Error()), "fault") {
		t.Fatalf("LnkSearchMachine returned a DCE/RPC fault (wire modeling bug): %v", err)
	}
}
