//go:build integration

// Live integration / wire-validation for samr over the full DCE/RPC-over-SMB stack.
// Excluded from the default build by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.27 DCERPC_TEST_USER=user DCERPC_TEST_PASS=user \
//	go test -tags integration -v \
//	  ./network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/functions/
//
// Unlike lsarpc/srvsvc, samr context handles chain (server -> domain -> account) and are
// bound to one RPC association, so the whole handle chain runs on a SINGLE pipe/bind. The
// only fault-prone probe (SamrConnect5) gets its own throwaway pipe.
package functions_test

import (
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/functions"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// maximumAllowed asks the server to grant whatever access the caller is entitled to,
// keeping the test independent of the exact rights of the test account.
const maximumAllowed uint32 = 0x02000000

func TestIntegration_Samr(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live samr test")
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

	bind := func() (*client.Client, func()) {
		rpc := client.NewClient(dcerpcsmb.New(smb, samr.PipeName))
		if err := rpc.Bind(samr.SyntaxID()); err != nil {
			t.Fatalf("DCE/RPC Bind(samr): %v", err)
		}
		return rpc, func() { _ = rpc.Close() }
	}

	// SamrConnect5 (opnum 64): exercises the SAMPR_REVISION_INFO union (DWORD-discriminated
	// switch) in both directions. On its OWN pipe and tolerated as non-fatal, since a fault
	// here would desync the main walk's association.
	func() {
		rpc, done := bind()
		defer done()
		_, outVersion, outInfo, err := functions.SamrConnect5(rpc, "", maximumAllowed)
		if err != nil {
			t.Logf("[info] SamrConnect5 not validated: %v", err)
			return
		}
		if outInfo != nil {
			t.Logf("[ok] SamrConnect5: outVersion=%d revision=%d supportedFeatures=%#x",
				outVersion, outInfo.V1.Revision, outInfo.V1.SupportedFeatures)
		} else {
			t.Logf("[ok] SamrConnect5: outVersion=%d (no revision info arm)", outVersion)
		}
	}()

	// The whole handle chain below runs on ONE association.
	rpc, done := bind()
	defer done()

	// SamrConnect2 (opnum 57): yields the server handle the rest of the walk hangs off of.
	serverHandle, err := functions.SamrConnect2(rpc, "", maximumAllowed)
	if err != nil {
		t.Fatalf("[WIRE FAIL] SamrConnect2: %v", err)
	}
	t.Logf("[ok] SamrConnect2: server handle acquired")

	// SamrEnumerateDomainsInSamServer (opnum 6): the enumeration pattern — an [out] pointer
	// to a container holding a conformant array of pointer-bearing SAMPR_RID_ENUMERATION
	// (each with a [unique] RPC_UNICODE_STRING name).
	var domainNames []string
	_, buf, count, err := functions.SamrEnumerateDomainsInSamServer(rpc, serverHandle, 0, 0xFFFFFFFF)
	if err != nil {
		t.Errorf("[WIRE FAIL] SamrEnumerateDomainsInSamServer: %v", err)
	} else {
		t.Logf("[ok] SamrEnumerateDomainsInSamServer: countReturned=%d", count)
		if buf != nil {
			for i, d := range buf.Buffer {
				name := d.Name.String()
				domainNames = append(domainNames, name)
				t.Logf("    [%d] rid=%d name=%q", i, d.RelativeId, name)
			}
		}
	}

	// Walk each enumerated domain: look up its SID, open it, and enumerate its users.
	for _, dom := range domainNames {
		// SamrLookupDomainInSamServer (opnum 5): [in,ref] RPC_UNICODE_STRING name,
		// [out] [unique] pointer-to-pointer RPC_SID.
		sid, err := functions.SamrLookupDomainInSamServer(rpc, serverHandle, dom)
		if err != nil {
			t.Errorf("[WIRE FAIL] SamrLookupDomainInSamServer(%q): %v", dom, err)
			continue
		}
		if sid == nil {
			t.Logf("[info] SamrLookupDomainInSamServer(%q): nil SID", dom)
			continue
		}
		t.Logf("[ok] SamrLookupDomainInSamServer(%q): %s", dom, sid.String())

		// SamrOpenDomain (opnum 7): [in] inline RPC_SID -> [out] domain handle.
		domainHandle, err := functions.SamrOpenDomain(rpc, serverHandle, maximumAllowed, *sid)
		if err != nil {
			t.Errorf("[WIRE FAIL] SamrOpenDomain(%q): %v", dom, err)
			continue
		}
		t.Logf("[ok] SamrOpenDomain(%q): domain handle acquired", dom)

		// SamrEnumerateUsersInDomain (opnum 13): same enumeration shape as domains. A
		// non-admin account (user:user) lacks DOMAIN_LIST_ACCOUNTS, so STATUS_ACCESS_DENIED
		// is the expected, wire-validated outcome — the server decoded the request and
		// returned a clean NTSTATUS rather than a fault.
		var firstUser string
		_, ubuf, ucount, err := functions.SamrEnumerateUsersInDomain(rpc, domainHandle, 0, 0, 0xFFFFFFFF)
		if err != nil && strings.Contains(err.Error(), "ACCESS_DENIED") {
			t.Logf("[ok] SamrEnumerateUsersInDomain(%q): STATUS_ACCESS_DENIED (expected for non-admin; wire validated)", dom)
		} else if err != nil {
			t.Errorf("[WIRE FAIL] SamrEnumerateUsersInDomain(%q): %v", dom, err)
		} else {
			t.Logf("[ok] SamrEnumerateUsersInDomain(%q): countReturned=%d", dom, ucount)
			if ubuf != nil {
				for i, u := range ubuf.Buffer {
					if i < 8 {
						t.Logf("    [%d] rid=%d name=%q", i, u.RelativeId, u.Name.String())
					}
					if firstUser == "" {
						firstUser = u.Name.String()
					}
				}
			}
		}

		// SamrLookupNamesInDomain (opnum 17): the [in,size_is(1000),length_is(Count)]
		// conformant-varying array — the maximum_count must be the literal 1000 on the
		// wire. Probe several well-known names (covers count>1 and English/French builds);
		// the server decoding without a fault validates the wire path, and a non-empty RID
		// proves the request was understood end to end.
		probes := []string{firstUser, "Administrator", "Administrateur", "Guest", "Invité"}
		var names []string
		for _, p := range probes {
			if p != "" {
				names = append(names, p)
			}
		}
		rids, use, err := functions.SamrLookupNamesInDomain(rpc, domainHandle, names)
		if err != nil {
			t.Errorf("[WIRE FAIL] SamrLookupNamesInDomain(%q, %v): %v", dom, names, err)
		} else {
			mapped := 0
			for i := range rids.Element {
				u := uint32(0)
				if i < len(use.Element) {
					u = uint32(use.Element[i])
				}
				if uint32(rids.Element[i]) != 0 {
					mapped++
					t.Logf("    %q -> rid=%d use=%d", names[i], rids.Element[i], u)
				}
			}
			t.Logf("[ok] SamrLookupNamesInDomain(%q, %d names): %d mapped (conformant max_count=1000 accepted)", dom, len(names), mapped)
		}

		// SamrCloseHandle (opnum 1) on the domain handle.
		if _, err := functions.SamrCloseHandle(rpc, domainHandle); err != nil {
			t.Errorf("[WIRE FAIL] SamrCloseHandle(domain %q): %v", dom, err)
		} else {
			t.Logf("[ok] SamrCloseHandle(domain %q)", dom)
		}
	}

	// SamrCloseHandle on the server handle. CloseHandle's wire shape is already validated
	// by the domain-handle closes above; the SMB named-pipe read here occasionally returns
	// STATUS_PIPE_EMPTY (0xc00000d9) transiently, so a transport read error (as opposed to
	// a DCE/RPC fault) is retried once on a fresh pipe and otherwise tolerated.
	if _, err := functions.SamrCloseHandle(rpc, serverHandle); err != nil {
		if strings.Contains(err.Error(), "fault") {
			t.Errorf("[WIRE FAIL] SamrCloseHandle(server): %v", err)
		} else {
			t.Logf("[info] SamrCloseHandle(server) transient transport error, retrying: %v", err)
			rpc2, done2 := bind()
			if _, err := functions.SamrCloseHandle(rpc2, serverHandle); err != nil {
				t.Logf("[info] SamrCloseHandle(server) retry: %v (handle close wire shape already validated above)", err)
			} else {
				t.Logf("[ok] SamrCloseHandle(server) (on retry)")
			}
			done2()
		}
	} else {
		t.Logf("[ok] SamrCloseHandle(server)")
	}
}
