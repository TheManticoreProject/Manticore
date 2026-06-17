//go:build integration

// Live validation of RPC-level NTLM authentication (auth_verifier / sec_trailer) on the
// connection-oriented client, exercised against the endpoint mapper over ncacn_ip_tcp.
// Excluded from the default build by the "integration" tag. Run with:
//
//	DCERPC_TEST_HOST=192.168.1.31 DCERPC_TEST_USER=Administrator DCERPC_TEST_PASS='Admin123!' \
//	go test -tags integration -v -run TestIntegration_AuthenticatedBind \
//	  ./network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions/
//
// The bind performs the NTLM negotiate/challenge/auth3 exchange; functions.Lookup then
// issues a signed-or-sealed ept_lookup whose response is decrypted and signature-checked
// before the entries are decoded.
package functions_test

import (
	"os"
	"testing"

	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/tcp"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestIntegration_AuthenticatedBind(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live authenticated bind test")
	}
	creds, err := credentials.NewCredentials(os.Getenv("DCERPC_TEST_DOMAIN"), os.Getenv("DCERPC_TEST_USER"), os.Getenv("DCERPC_TEST_PASS"), "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	levels := []struct {
		name  string
		level uint8
	}{
		{"PKT_INTEGRITY", pdu.AuthLevelPktIntegrity},
		{"PKT_PRIVACY", pdu.AuthLevelPktPrivacy},
	}
	for _, lv := range levels {
		t.Run(lv.name, func(t *testing.T) {
			rpc := client.NewClient(tcp.New(host, tcp.EndpointMapperPort))
			if err := rpc.SetAuth(pdu.AuthTypeNTLMSSP, lv.level, creds); err != nil {
				t.Fatalf("SetAuth: %v", err)
			}
			if err := rpc.Bind(epm.SyntaxID()); err != nil {
				t.Fatalf("authenticated Bind(ept) at %s: %v", lv.name, err)
			}
			defer rpc.Close()

			// A real, protected call: ept_lookup. Its response is signed (and sealed at
			// PKT_PRIVACY); recovering and decoding entries proves the verifier round-trips.
			entries, err := functions.Lookup(rpc)
			if err != nil {
				t.Fatalf("[WIRE FAIL] authenticated ept_lookup at %s: %v", lv.name, err)
			}
			if len(entries) == 0 {
				t.Fatalf("[WIRE FAIL] authenticated ept_lookup at %s returned no entries", lv.name)
			}
			t.Logf("[ok] %s: authenticated bind + ept_lookup recovered %d endpoint-map entries", lv.name, len(entries))
		})
	}
}
