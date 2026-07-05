//go:build integration

// Live integration / wire-validation for winreg (MS-RRP). Excluded from the default build
// by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.30 DCERPC_TEST_USER=Administrator DCERPC_TEST_PASS=admin \
//	go test -tags integration -v \
//	  ./network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/functions/
//
// The test opens HKEY_LOCAL_MACHINE (OpenLocalMachine), reads the protocol version
// (BaseRegGetVersion), opens a well-known subkey (BaseRegOpenKey), queries its key info
// (BaseRegQueryInfoKey), enumerates its first subkey (BaseRegEnumKey), and closes both
// handles (BaseRegCloseKey). Registry context handles chain (HKLM -> subkey -> close), so
// the whole sequence runs on ONE association.
//
// Per [MS-RRP] every method returns ERROR_SUCCESS (0) or a Win32 error ([MS-ERREF]); a
// clean nonzero status still proves the wire is correct, so only a DCE/RPC fault
// (nca_s_fault_*) is treated as a wire failure here.
package functions_test

import (
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/functions"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// regName builds an RRP_UNICODE_STRING for a key/value name, including the terminating NUL
// in the counted length as [MS-RRP] 3.1.1 requires.
func regName(s string) msrrp.RRP_UNICODE_STRING { return dtyp.NewUnicodeString(s + "\x00") }

// isFault reports whether err is a DCE/RPC fault (a wire-modeling failure) rather than a
// clean Win32 error returned by the server (which validates the wire).
func isFault(err error) bool {
	return err != nil && strings.Contains(err.Error(), "fault")
}

func TestIntegration_Winreg(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live winreg test")
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

	// Registry handles chain, so bind once and run the whole sequence on it.
	rpc := client.NewClient(dcerpcsmb.New(smb, winreg.PipeName))
	if err := rpc.Bind(winreg.SyntaxID()); err != nil {
		t.Fatalf("Bind(winreg): %v", err)
	}
	defer rpc.Close()

	// OpenLocalMachine (opnum 2): nil ServerName -> the bound host's HKLM.
	hklm, err := functions.OpenLocalMachine(rpc, nil, winreg.KeyRead)
	if err != nil {
		if isFault(err) {
			t.Fatalf("[WIRE FAIL] OpenLocalMachine: %v", err)
		}
		t.Fatalf("OpenLocalMachine: server returned status: %v", err)
	}
	t.Logf("[ok] OpenLocalMachine: HKLM handle acquired")

	// BaseRegGetVersion (opnum 26): a single [out] DWORD scalar.
	if ver, err := functions.BaseRegGetVersion(rpc, hklm); err != nil {
		if isFault(err) {
			t.Errorf("[WIRE FAIL] BaseRegGetVersion: %v", err)
		} else {
			t.Logf("[ok] BaseRegGetVersion: status (wire validated): %v", err)
		}
	} else {
		t.Logf("[ok] BaseRegGetVersion: registry version %d", uint32(ver))
	}

	// BaseRegOpenKey (opnum 15): inline RRP_UNICODE_STRING subkey, chains a subkey handle.
	const subkey = `SOFTWARE\Microsoft\Windows NT\CurrentVersion`
	hkey, err := functions.BaseRegOpenKey(rpc, hklm, regName(subkey), 0, winreg.KeyRead)
	if err != nil {
		if isFault(err) {
			t.Fatalf("[WIRE FAIL] BaseRegOpenKey: %v", err)
		}
		t.Logf("[ok] BaseRegOpenKey(%s): status (wire validated): %v", subkey, err)
	} else {
		t.Logf("[ok] BaseRegOpenKey(%s): subkey handle acquired", subkey)

		// BaseRegQueryInfoKey (opnum 16): inline class in/out, eight [out] DWORD scalars,
		// and a [out] FILETIME.
		if _, nSub, _, _, nVal, _, _, _, ft, err := functions.BaseRegQueryInfoKey(rpc, hkey, dtyp.RPC_UNICODE_STRING{}); err != nil {
			if isFault(err) {
				t.Errorf("[WIRE FAIL] BaseRegQueryInfoKey: %v", err)
			} else {
				t.Logf("[ok] BaseRegQueryInfoKey: status (wire validated): %v", err)
			}
		} else {
			t.Logf("[ok] BaseRegQueryInfoKey: %d subkeys, %d values, lastWrite=%#x", uint32(nSub), uint32(nVal), ft.Uint64())
		}

		// BaseRegEnumKey (opnum 9): exercises the [in] sized name buffer, the [out,unique]
		// double-pointer class, and the [in,out,unique] FILETIME pointer.
		nameIn := msrrp.RRP_UNICODE_STRING{Length: 0, MaximumLength: 512, Buffer: make([]uint16, 256)}
		classIn := msrrp.RRP_UNICODE_STRING{Length: 0, MaximumLength: 512, Buffer: make([]uint16, 256)}
		if nameOut, _, _, err := functions.BaseRegEnumKey(rpc, hkey, 0, nameIn, &classIn, nil); err != nil {
			if isFault(err) {
				t.Errorf("[WIRE FAIL] BaseRegEnumKey: %v", err)
			} else {
				t.Logf("[ok] BaseRegEnumKey: status (wire validated): %v", err)
			}
		} else {
			t.Logf("[ok] BaseRegEnumKey[0]: %q", nameOut.String())
		}

		// BaseRegCloseKey (opnum 5): [in,out] handle round-trip.
		if _, err := functions.BaseRegCloseKey(rpc, hkey); err != nil && isFault(err) {
			t.Errorf("[WIRE FAIL] BaseRegCloseKey(subkey): %v", err)
		} else {
			t.Logf("[ok] BaseRegCloseKey(subkey)")
		}
	}

	if _, err := functions.BaseRegCloseKey(rpc, hklm); err != nil && isFault(err) {
		t.Errorf("[WIRE FAIL] BaseRegCloseKey(HKLM): %v", err)
	} else {
		t.Logf("[ok] BaseRegCloseKey(HKLM)")
	}
}
