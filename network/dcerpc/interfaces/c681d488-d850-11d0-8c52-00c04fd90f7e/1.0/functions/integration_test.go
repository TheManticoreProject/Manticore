//go:build integration

// Live integration / wire-validation for efsrpc (MS-EFSR), including the NDR-pipe
// raw-file methods. Excluded from the default build by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.30 DCERPC_TEST_USER=Administrator DCERPC_TEST_PASS=admin \
//	go test -tags integration -v \
//	  ./network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/functions/
//
// The test stages a file on C$, encrypts it server-side (EfsRpcEncryptFileSrv), then
// exports it raw through the [out] EFS_EXIM_PIPE (EfsRpcOpenFileRaw -> ReadFileRaw ->
// CloseRaw). EFSRPC context handles chain, so that sequence runs on one association.
//
// Per [MS-EFSR] every method returns 0 on success or a Win32 error (MS-ERREF) on
// failure; a clean nonzero status still proves the wire is correct, so only a DCE/RPC
// fault (nca_s_fault_*) is treated as a wire failure here.
package functions_test

import (
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0/functions"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

func ws(s string) ndr.WSTR { return ndr.WSTR(s) }

// isFault reports whether err is a DCE/RPC fault (a wire-modeling failure) rather than a
// clean Win32/NTSTATUS error returned by the server (which validates the wire).
func isFault(err error) bool {
	return err != nil && strings.Contains(err.Error(), "fault")
}

func TestIntegration_Efsr(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live efsr test")
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

	// --- Stage a file on C$ so the raw export has real data to stream. ---
	const sharePath = `\WINDOWS\Temp\efsr_pipe_test.txt` // C$-relative
	const serverPath = `C:\WINDOWS\Temp\efsr_pipe_test.txt`
	payload := []byte("EFSR pipe streaming live test — " + strings.Repeat("ABCD", 64))
	staged := false
	func() {
		if err := smb.TreeConnect("C$"); err != nil {
			t.Logf("[info] TreeConnect(C$) failed (%v); skipping file staging", err)
			return
		}
		defer smb.TreeDisconnect()
		fid, err := smb.OpenFile(sharePath, fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
			fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OVERWRITE_IF, fileflags.FILE_NON_DIRECTORY_FILE)
		if err != nil {
			t.Logf("[info] OpenFile(%s) failed: %v", sharePath, err)
			return
		}
		if _, err := smb.WriteFile(fid, 0, payload); err != nil {
			t.Logf("[info] WriteFile failed: %v", err)
			_ = smb.CloseFile(fid)
			return
		}
		_ = smb.CloseFile(fid)
		staged = true
		t.Logf("[ok] staged %d-byte %s on C$", len(payload), serverPath)
	}()

	if err := smb.TreeConnect("IPC$"); err != nil {
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	defer smb.TreeDisconnect()

	pipeName := chooseEfsrPipe(t, smb)
	t.Logf("[ok] bound efsrpc over %s", pipeName)

	bind := func() (*client.Client, func()) {
		rpc := client.NewClient(dcerpcsmb.New(smb, pipeName))
		if err := rpc.Bind(efsrpc.SyntaxID()); err != nil {
			t.Fatalf("Bind(efsrpc): %v", err)
		}
		return rpc, func() { _ = rpc.Close() }
	}

	// EfsRpcEncryptFileSrv (opnum 4): encrypt the staged file server-side. Validates the
	// [in,string] FileName parameter ([ref] inline wide string).
	if staged {
		rpc, done := bind()
		err := functions.EfsRpcEncryptFileSrv(rpc, ws(serverPath))
		switch {
		case isFault(err):
			t.Errorf("[WIRE FAIL] EfsRpcEncryptFileSrv: %v", err)
		case err != nil:
			t.Logf("[ok] EfsRpcEncryptFileSrv: server returned a status (wire validated): %v", err)
		default:
			t.Logf("[ok] EfsRpcEncryptFileSrv(%s): encrypted", serverPath)
		}
		done()
	}

	// Raw-export handle chain on ONE association.
	func() {
		rpc, done := bind()
		defer done()
		hContext, err := functions.EfsRpcOpenFileRaw(rpc, ws(serverPath), 0)
		switch {
		case isFault(err):
			t.Errorf("[WIRE FAIL] EfsRpcOpenFileRaw: %v", err)
			return
		case err != nil:
			t.Logf("[ok] EfsRpcOpenFileRaw: server returned a status (wire validated): %v", err)
			return // no context handle -> cannot read/close
		}
		t.Logf("[ok] EfsRpcOpenFileRaw: context handle acquired")

		pipeData, err := functions.EfsRpcReadFileRaw(rpc, hContext)
		switch {
		case isFault(err):
			t.Errorf("[WIRE FAIL] EfsRpcReadFileRaw (the [out] NDR pipe): %v", err)
		case err != nil:
			t.Logf("[ok] EfsRpcReadFileRaw: status (wire validated): %v", err)
		default:
			head := pipeData
			if len(head) > 16 {
				head = head[:16]
			}
			t.Logf("[ok] EfsRpcReadFileRaw: streamed %d bytes via the pipe; head=%x", len(pipeData), head)
		}

		if _, err := functions.EfsRpcCloseRaw(rpc, hContext); isFault(err) {
			t.Errorf("[WIRE FAIL] EfsRpcCloseRaw: %v", err)
		} else {
			t.Logf("[ok] EfsRpcCloseRaw (err=%v)", err)
		}
	}()

	// EfsRpcQueryUsersOnFile (opnum 6): another [in,string] method.
	func() {
		rpc, done := bind()
		defer done()
		users, err := functions.EfsRpcQueryUsersOnFile(rpc, ws(serverPath))
		switch {
		case isFault(err):
			t.Errorf("[WIRE FAIL] EfsRpcQueryUsersOnFile: %v", err)
		case err != nil:
			t.Logf("[ok] EfsRpcQueryUsersOnFile: status (wire validated): %v", err)
		default:
			t.Logf("[ok] EfsRpcQueryUsersOnFile: users=%v", users != nil)
		}
	}()
}

// chooseEfsrPipe binds efsrpc, trying \efsrpc then \lsarpc on a throwaway client.
func chooseEfsrPipe(t *testing.T, smb *smbclient.Client) string {
	t.Helper()
	for _, name := range []string{`\efsrpc`, `\lsarpc`} {
		rpc := client.NewClient(dcerpcsmb.New(smb, name))
		err := rpc.Bind(efsrpc.SyntaxID())
		_ = rpc.Close()
		if err == nil {
			return name
		}
		t.Logf("[info] bind over %s failed: %v", name, err)
	}
	t.Fatalf("could not bind efsrpc over \\efsrpc or \\lsarpc")
	return ""
}
