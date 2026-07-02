//go:build integration

// Live-DC wire-validation harness for the lsarpc methods whose NDR is more complex than
// the basic OpenPolicy2/Close path (issue #421 follow-up). It exercises only READ-ONLY
// methods (Query/Lookup/Enumerate) — it never creates, sets, or deletes server state.
//
// For each method it classifies the outcome:
//   - decoded OK   : the call returned STATUS_SUCCESS and the response NDR decoded;
//   - server status: the server returned an NTSTATUS (e.g. ACCESS_DENIED) — the request
//     was understood and the status decoded, so the wire path is valid;
//   - WIRE FAIL    : a transport/fault/NDR error — the request marshalling or response
//     decoding is wrong. This is the only outcome that fails the test.
//
// Run with:
//
//	DCERPC_TEST_HOST=192.168.1.27 DCERPC_TEST_USER=user DCERPC_TEST_PASS=user \
//	go test -tags integration -run TestLiveValidation -v \
//	  ./network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/functions/
package functions_test

import (
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
	mslsat "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsat"
)

// classify reports the wire outcome of a method call:
//   - nil error                 -> decoded OK (STATUS_SUCCESS, response NDR decoded);
//   - transport/pipe/fault error -> WIRE FAIL (request marshalling or response decode
//     wrong; the server faulted or tore down the pipe);
//   - "<Method> failed: <ntstatus>" -> the server returned an application NTSTATUS, i.e.
//     the request was understood and the status decoded.
//
// Returns (wireOK, decodedSuccess).
func classify(t *testing.T, method string, err error) (bool, bool) {
	t.Helper()
	if err == nil {
		t.Logf("[decoded OK]    %s", method)
		return true, true
	}
	msg := err.Error()
	transport := strings.Contains(msg, "named pipe") || strings.Contains(msg, "ReadAndx") ||
		strings.Contains(msg, "dcerpc call") || strings.Contains(msg, "fault") ||
		strings.Contains(msg, "unmarshal") || strings.Contains(msg, "decode")
	// Our method wrappers report a decoded NTSTATUS as "<Method> failed: <status>".
	appStatus := strings.Contains(msg, "failed: STATUS_") || strings.Contains(msg, "failed: 0x")
	if appStatus && !transport {
		t.Logf("[server status] %s -> %v", method, err)
		return true, false
	}
	t.Errorf("[WIRE FAIL]     %s -> %v", method, err)
	return false, false
}

func TestLiveValidation(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live validation")
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

	// withPolicy runs fn on a FRESH pipe + bind + policy handle, so a method that faults
	// or tears down the pipe cannot corrupt the next method's result.
	withPolicy := func(name string, fn func(rpc *client.Client, policy mslsad.LSAPR_HANDLE)) {
		rpc := client.NewClient(dcerpcsmb.New(smb, lsarpc.PipeName))
		defer rpc.Close()
		if err := rpc.Bind(lsarpc.SyntaxID()); err != nil {
			t.Errorf("[WIRE FAIL]     %s: bind: %v", name, err)
			return
		}
		policy, err := functions.LsarOpenPolicy2(rpc, lsarpc.MaximumAllowed)
		if err != nil {
			t.Errorf("[setup fail]    %s: OpenPolicy2: %v", name, err)
			return
		}
		defer func() { _, _ = functions.LsarClose(rpc, policy) }()
		fn(rpc, policy)
	}

	// --- Union path: LsarQueryInformationPolicy (union decode + RPC_UNICODE_STRING + SID) ---
	withPolicy("LsarQueryInformationPolicy(AccountDomain)", func(rpc *client.Client, policy mslsad.LSAPR_HANDLE) {
		info, err := functions.LsarQueryInformationPolicy(rpc, policy, mslsad.PolicyAccountDomainInformation)
		if _, ok := classify(t, "LsarQueryInformationPolicy(AccountDomain)", err); ok && err == nil {
			ad := info.PolicyAccountDomainInfo
			sid := "<nil>"
			if ad.DomainSid != nil {
				sid = ad.DomainSid.String()
			}
			t.Logf("    account domain: name=%q sid=%s", ad.DomainName.String(), sid)
		}
	})
	withPolicy("LsarQueryInformationPolicy(PrimaryDomain)", func(rpc *client.Client, policy mslsad.LSAPR_HANDLE) {
		info, err := functions.LsarQueryInformationPolicy(rpc, policy, mslsad.PolicyPrimaryDomainInformation)
		if _, ok := classify(t, "LsarQueryInformationPolicy(PrimaryDomain)", err); ok && err == nil {
			pd := info.PolicyPrimaryDomainInfo
			sid := "<nil>"
			if pd.Sid != nil {
				sid = pd.Sid.String()
			}
			t.Logf("    primary domain: name=%q sid=%s", pd.Name.String(), sid)
		}
	})

	// Isolation probe: a union arm with NO RPC_UNICODE_STRING (server-role = a bare enum),
	// to tell whether union failures are the union machinery or the string type.
	withPolicy("LsarQueryInformationPolicy(ServerRole)", func(rpc *client.Client, policy mslsad.LSAPR_HANDLE) {
		info, err := functions.LsarQueryInformationPolicy(rpc, policy, mslsad.PolicyLsaServerRoleInformation)
		if _, ok := classify(t, "LsarQueryInformationPolicy(ServerRole)", err); ok && err == nil {
			t.Logf("    server role = %d", info.PolicyServerRoleInfo.LsaServerRole)
		}
	})

	// --- SID->name: LsarLookupSids (SID array in + translated-names array out + referenced domains) ---
	withPolicy("LsarLookupSids", func(rpc *client.Client, policy mslsad.LSAPR_HANDLE) {
		var sidInfos []mslsat.LSAPR_SID_INFORMATION
		for _, s := range []string{"S-1-5-32-544", "S-1-1-0", "S-1-5-18"} {
			sid, perr := dtyp.ParseSID(s)
			if perr != nil {
				t.Fatalf("ParseSID(%s): %v", s, perr)
			}
			sidInfos = append(sidInfos, mslsat.LSAPR_SID_INFORMATION{Sid: &sid})
		}
		sidBuf := mslsat.LSAPR_SID_ENUM_BUFFER{Entries: uint32(len(sidInfos)), SidInfo: sidInfos}
		dom, names, mapped, err := functions.LsarLookupSids(rpc, policy, sidBuf, mslsat.LsapLookupWksta)
		if _, ok := classify(t, "LsarLookupSids(Administrators,Everyone,System)", err); ok && err == nil {
			t.Logf("    mapped=%d names.Entries=%d", mapped, names.Entries)
			for i, n := range names.Names {
				d := "?"
				if dom != nil && int(n.DomainIndex) >= 0 && int(n.DomainIndex) < len(dom.Domains) {
					d = dom.Domains[n.DomainIndex].Name.String()
				}
				t.Logf("    [%d] use=%d name=%q domain=%q", i, n.Use, n.Name.String(), d)
			}
		}
	})

	// --- name->SID: LsarLookupNames (RPC_UNICODE_STRING array in + translated-sids out) ---
	withPolicy("LsarLookupNames", func(rpc *client.Client, policy mslsad.LSAPR_HANDLE) {
		names := []dtyp.RPC_UNICODE_STRING{dtyp.NewUnicodeString("Administrator"), dtyp.NewUnicodeString("Guest")}
		dom, sids, mapped, err := functions.LsarLookupNames(rpc, policy, names, mslsat.LsapLookupWksta)
		if _, ok := classify(t, "LsarLookupNames(Administrator,Guest)", err); ok && err == nil {
			t.Logf("    mapped=%d sids.Entries=%d referencedDomains=%d", mapped, sids.Entries, domCount(dom))
			for i, s := range sids.Sids {
				t.Logf("    [%d] use=%d rid=%d domainIndex=%d", i, s.Use, s.RelativeId, s.DomainIndex)
			}
		}
	})

	// --- LUID + string: LsarLookupPrivilegeValue then LsarLookupPrivilegeName ---
	withPolicy("LsarLookupPrivilegeValue/Name", func(rpc *client.Client, policy mslsad.LSAPR_HANDLE) {
		luid, err := functions.LsarLookupPrivilegeValue(rpc, policy, "SeShutdownPrivilege")
		if _, ok := classify(t, "LsarLookupPrivilegeValue(SeShutdownPrivilege)", err); ok && err == nil {
			t.Logf("    LUID = {low=%#x high=%#x}", luid.LowPart, luid.HighPart)
			name, err := functions.LsarLookupPrivilegeName(rpc, policy, luid)
			if _, ok := classify(t, "LsarLookupPrivilegeName(<luid>)", err); ok && err == nil && name != nil {
				t.Logf("    name = %q", name.String())
			}
		}
	})

	// --- Array-of-pointer-struct buffers ---
	withPolicy("LsarEnumeratePrivileges", func(rpc *client.Client, policy mslsad.LSAPR_HANDLE) {
		_, buf, err := functions.LsarEnumeratePrivileges(rpc, policy, 0, 0x10000)
		if _, ok := classify(t, "LsarEnumeratePrivileges", err); ok && err == nil {
			t.Logf("    privileges returned: %d", buf.Entries)
			for i, p := range buf.Privileges {
				if i >= 5 {
					t.Logf("    ... (%d total)", buf.Entries)
					break
				}
				t.Logf("    [%d] %q luid={low=%#x high=%#x}", i, p.Name.String(), p.LocalValue.LowPart, p.LocalValue.HighPart)
			}
		}
	})
	withPolicy("LsarEnumerateAccounts", func(rpc *client.Client, policy mslsad.LSAPR_HANDLE) {
		buf, _, err := functions.LsarEnumerateAccounts(rpc, policy, 0, 0x10000)
		if _, ok := classify(t, "LsarEnumerateAccounts", err); ok && err == nil {
			t.Logf("    accounts returned: %d", buf.EntriesRead)
			for i, a := range buf.Information {
				if i >= 8 {
					t.Logf("    ... (%d total)", buf.EntriesRead)
					break
				}
				sid := "<nil>"
				if a.Sid != nil {
					sid = a.Sid.String()
				}
				t.Logf("    [%d] %s", i, sid)
			}
		}
	})
}

func domCount(d *mslsat.LSAPR_REFERENCED_DOMAIN_LIST) uint32 {
	if d == nil {
		return 0
	}
	return d.Entries
}
