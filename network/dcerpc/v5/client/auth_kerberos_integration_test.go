//go:build integration

// Live validation of RPC-level Kerberos authentication (SetAuthKerberos) on the
// connection-oriented DCE/RPC client, exercised against the endpoint mapper (ept)
// over ncacn_ip_tcp. It mirrors the NTLM auth_integration_test precedent but
// drives the native Kerberos provider: an AP-REQ in the bind's auth verifier,
// mutual authentication, and per-PDU GSS protection (MIC for PKT/PKT_INTEGRITY,
// Wrap for PKT_PRIVACY). Excluded from the default build by the "integration" tag.
//
// Configuration:
//
//	KRB5_TEST_KDC     KDC / domain-controller host or IP (obtains the TGT)
//	KRB5_TEST_REALM   Kerberos realm / AD domain
//	KRB5_TEST_USER    account sAMAccountName
//	KRB5_TEST_PASS    account password
//	KRB5_TEST_TARGET  RPC server FQDN (defaults to KRB5_TEST_KDC). The SPN is
//	                  host/<target> (the RPC/host class); the transport dials the
//	                  resolved host on TCP 135 (the endpoint mapper).
//
// Example:
//
//	KRB5_TEST_KDC=10.0.0.10 KRB5_TEST_REALM=EXAMPLE.LOCAL \
//	KRB5_TEST_USER=Administrator KRB5_TEST_PASS='…' \
//	KRB5_TEST_TARGET=dc.example.local \
//	  go test -tags integration -v -run TestLiveDCERPC ./network/dcerpc/v5/client/
package client_test

import (
	"os"
	"testing"

	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/tcp"
	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
)

// TestLiveDCERPC_AuthKerberos_EptLookup binds the endpoint mapper with native
// Kerberos at each supported auth level and issues a protected ept_lookup. The
// response is signature-checked (and decrypted at PKT_PRIVACY) before its entries
// are decoded, so a successful lookup proves the per-PDU verifier round-trips.
func TestLiveDCERPC_AuthKerberos_EptLookup(t *testing.T) {
	kdc := os.Getenv("KRB5_TEST_KDC")
	realm := os.Getenv("KRB5_TEST_REALM")
	user := os.Getenv("KRB5_TEST_USER")
	pass := os.Getenv("KRB5_TEST_PASS")
	if kdc == "" || realm == "" || user == "" || pass == "" {
		t.Skip("set KRB5_TEST_KDC/KRB5_TEST_REALM/KRB5_TEST_USER/KRB5_TEST_PASS to run the live DCE/RPC Kerberos tests")
	}
	target := os.Getenv("KRB5_TEST_TARGET")
	if target == "" {
		target = kdc
	}
	spn := "host/" + target

	levels := []struct {
		name  string
		level uint8
	}{
		{"CONNECT", pdu.AuthLevelConnect},
		{"PKT", pdu.AuthLevelPkt},
		{"PKT_INTEGRITY", pdu.AuthLevelPktIntegrity},
		{"PKT_PRIVACY", pdu.AuthLevelPktPrivacy},
	}
	for _, lv := range levels {
		t.Run(lv.name, func(t *testing.T) {
			// A fresh Kerberos client per level; SetAuthKerberos acquires the TGT
			// and the host/ service ticket, then builds the AP-REQ for the bind.
			kc := kerberos.NewClient(user, realm, kdc).WithPassword(pass)

			rpc := client.NewClient(tcp.New(target, tcp.EndpointMapperPort))
			if err := rpc.SetAuthKerberos(lv.level, kc, spn); err != nil {
				t.Fatalf("SetAuthKerberos(%s): %v", lv.name, err)
			}
			if err := rpc.Bind(epm.SyntaxID()); err != nil {
				t.Fatalf("authenticated Bind(ept) at %s: %v", lv.name, err)
			}
			defer rpc.Close()

			entries, err := functions.Lookup(rpc)
			if err != nil {
				t.Fatalf("[WIRE FAIL] authenticated ept_lookup at %s: %v", lv.name, err)
			}
			if len(entries) == 0 {
				t.Fatalf("[WIRE FAIL] authenticated ept_lookup at %s returned no entries", lv.name)
			}
			t.Logf("[ok] %s: Kerberos bind + ept_lookup recovered %d endpoint-map entries", lv.name, len(entries))
		})
	}
}
