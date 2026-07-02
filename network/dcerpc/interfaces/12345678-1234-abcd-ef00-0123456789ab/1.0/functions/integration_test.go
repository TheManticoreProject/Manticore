//go:build integration

// Integration test for the MS-RPRN (Print System Remote Protocol) winspool interface
// against a live spooler.
//
// Excluded from the default build by the "integration" tag; it needs a reachable host
// running the Print Spooler service. Run it with, for example:
//
//	DCERPC_TEST_HOST=192.168.1.40 \
//	DCERPC_TEST_USER=Administrator DCERPC_TEST_PASS='Admin123!' \
//	go test -tags integration -v \
//	  ./network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/functions/
//
// Optional: DCERPC_TEST_DOMAIN, DCERPC_TEST_PORT (default 445, the SMB port).
//
// Transport: the harness first tries MS-RPRN over the \spoolss named pipe (SMB, any
// negotiated dialect). Modern/hardened spoolers no longer expose that pipe and serve
// MS-RPRN only over ncacn_ip_tcp, so on OBJECT_NAME_NOT_FOUND it falls back to resolving
// the interface through the endpoint mapper (TCP/135) and binding over TCP with NTLM
// packet privacy. Validated 2026-07-01 against Windows at 192.168.1.40 (SMB v2.0.2 host,
// spooler reached over ncacn_ip_tcp): EnumPrinters returned 1 printer, OpenPrinter/
// ClosePrinter and GetPrinterDriverDirectory succeeded.
package functions_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/msproto"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// bindSpooler binds the winspool interface against the target, preferring the \spoolss
// named pipe over SMB and falling back to ncacn_ip_tcp (endpoint-mapper resolved) when the
// pipe is absent. It returns the bound client as an ndr.Invoker.
func bindSpooler(t *testing.T) ndr.Invoker {
	t.Helper()
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
	creds, err := credentials.NewCredentials(domain, user, pass, "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	// Attempt 1: MS-RPRN over the \spoolss named pipe on an SMB session.
	smb, err := smbclient.Dial(host, port, smbclient.Options{})
	if err != nil {
		t.Fatalf("SMB Dial: %v", err)
	}
	smbClosed := false
	closeSMB := func() {
		if !smbClosed {
			smbClosed = true
			smb.TreeDisconnect()
			smb.Logoff()
			smb.Disconnect()
		}
	}
	if err := smb.Login(creds); err != nil {
		smb.Disconnect()
		t.Fatalf("SMB Login: %v", err)
	}
	t.Logf("negotiated SMB dialect %v", smb.Dialect())
	if err := smb.TreeConnect("IPC$"); err != nil {
		smb.Logoff()
		smb.Disconnect()
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	if pipe, err := smb.RPCTransport(winspool.PipeName); err == nil {
		rpc := client.NewClient(pipe)
		if err := rpc.Bind(winspool.SyntaxID()); err == nil {
			t.Log("bound to winspool over \\spoolss (named pipe)")
			t.Cleanup(func() { rpc.Close(); closeSMB() })
			return rpc
		} else if !strings.Contains(err.Error(), "OBJECT_NAME_NOT_FOUND") {
			closeSMB()
			t.Fatalf("Bind over \\spoolss: %v", err)
		}
	}
	// The spooler does not expose the named pipe on this host; drop the SMB session.
	closeSMB()
	t.Log("\\spoolss named pipe absent; falling back to ncacn_ip_tcp via endpoint mapper")

	// Attempt 2: resolve winspool through the endpoint mapper (TCP/135) and bind over TCP.
	rpc, closer, err := msproto.NewTCPBinder(host, 0, creds, 0).Bind(winspool.SyntaxID())
	if err != nil {
		t.Fatalf("bind winspool over ncacn_ip_tcp: %v", err)
	}
	t.Log("bound to winspool over ncacn_ip_tcp")
	t.Cleanup(func() { closer() })
	return rpc
}

// TestIntegration_EnumPrinters exercises the [in,out,unique] buffer two-call sizing
// pattern (opnum 0): a first call with a NULL buffer returns the required size, then a
// second call with the allocated buffer returns the packed PRINTER_INFO array.
func TestIntegration_EnumPrinters(t *testing.T) {
	rpc := bindSpooler(t)

	const printerEnumLocal ndr.DWORD = 0x00000002 // PRINTER_ENUM_LOCAL
	_, needed, returned, err := functions.RpcEnumPrinters(rpc, printerEnumLocal, nil, 1, nil, 0)
	t.Logf("RpcEnumPrinters(sizing): pcbNeeded=%d pcReturned=%d err=%v", needed, returned, err)
	if err != nil && needed == 0 {
		t.Fatalf("RpcEnumPrinters sizing call failed with no size hint: %v", err)
	}
	if needed == 0 {
		t.Log("server reports no local printers; buffer round-trip skipped")
		return
	}

	buf := make([]uint8, needed)
	out, needed2, returned2, err := functions.RpcEnumPrinters(rpc, printerEnumLocal, nil, 1, buf, needed)
	if err != nil {
		t.Fatalf("RpcEnumPrinters data call failed (needed=%d): %v", needed, err)
	}
	if len(out) != int(needed) {
		t.Errorf("returned buffer len=%d, want %d", len(out), needed)
	}
	t.Logf("RpcEnumPrinters(data): %d printer(s), pcbNeeded=%d buffer=%d bytes", returned2, needed2, len(out))
}

// TestIntegration_OpenClosePrinter exercises a PRINTER_HANDLE context-handle round-trip
// (opnums 1 and 29): open the print server object, then close the returned handle.
func TestIntegration_OpenClosePrinter(t *testing.T) {
	rpc := bindSpooler(t)

	server := ndr.WSTR(`\\` + os.Getenv("DCERPC_TEST_HOST"))
	empty := structures.DEVMODE_CONTAINER{} // CbBuf=0, PDevMode=nil

	var handle structures.PRINTER_HANDLE
	var err error
	// Try the access masks a print server accepts, most-permissive first; a server may
	// reject MAXIMUM_ALLOWED for the server object and require an explicit SERVER_* mask.
	for _, access := range []ndr.DWORD{0x000F0003 /*SERVER_ALL_ACCESS*/, 0x00020002 /*SERVER_READ*/, 0x02000000 /*MAXIMUM_ALLOWED*/} {
		handle, err = functions.RpcOpenPrinter(rpc, &server, nil, empty, access)
		if err == nil {
			t.Logf("RpcOpenPrinter(%s, access=0x%08x) -> handle %x", server, access, handle)
			break
		}
		t.Logf("RpcOpenPrinter access=0x%08x rejected: %v", access, err)
	}
	if err != nil {
		t.Fatalf("RpcOpenPrinter failed for all access masks: %v", err)
	}

	if _, err := functions.RpcClosePrinter(rpc, handle); err != nil {
		t.Fatalf("RpcClosePrinter: %v", err)
	}
	t.Log("printer handle closed")
}

// TestIntegration_GetPrinterDriverDirectory exercises another two-call sizing method
// (opnum 12) with a [unique] out buffer.
func TestIntegration_GetPrinterDriverDirectory(t *testing.T) {
	rpc := bindSpooler(t)

	_, needed, err := functions.RpcGetPrinterDriverDirectory(rpc, nil, nil, 1, nil, 0)
	t.Logf("RpcGetPrinterDriverDirectory(sizing): pcbNeeded=%d err=%v", needed, err)
	if err != nil && needed == 0 {
		t.Fatalf("RpcGetPrinterDriverDirectory sizing failed with no size hint: %v", err)
	}
	if needed == 0 {
		return
	}
	buf := make([]uint8, needed)
	out, needed2, err := functions.RpcGetPrinterDriverDirectory(rpc, nil, nil, 1, buf, needed)
	if err != nil {
		t.Fatalf("RpcGetPrinterDriverDirectory data call: %v", err)
	}
	t.Logf("RpcGetPrinterDriverDirectory(data): %d bytes (pcbNeeded=%d)", len(out), needed2)
}
