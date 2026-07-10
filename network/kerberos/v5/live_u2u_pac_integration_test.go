//go:build integration

// White-box live coverage for the two native-Kerberos paths that need access to
// the client's internal TGT state: a user-to-user (U2U) TGS exchange and PAC
// decoding of the resulting ticket. Both use the baseline KRB5_TEST_* config (see
// live_integration_test.go). Excluded from the default build by the "integration"
// tag.
//
// The U2U ticket to the account itself is encrypted under the client's own TGT
// session key (ENC-TKT-IN-SKEY), so the test can decrypt the ticket enc-part,
// recover the AD-IF-RELEVANT -> AD-WIN2K-PAC element, and decode the PAC — a
// genuine end-to-end PAC decode against a live KDC.
package kerberos

import (
	"encoding/asn1"
	"os"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pac"
)

// liveBaseline mirrors requireKrbEnv for the white-box package.
func liveBaseline(t *testing.T) (kdc, realm, user, pass string) {
	t.Helper()
	kdc = os.Getenv("KRB5_TEST_KDC")
	realm = os.Getenv("KRB5_TEST_REALM")
	user = os.Getenv("KRB5_TEST_USER")
	pass = os.Getenv("KRB5_TEST_PASS")
	if kdc == "" || realm == "" || user == "" || pass == "" {
		t.Skip("set KRB5_TEST_KDC/KRB5_TEST_REALM/KRB5_TEST_USER/KRB5_TEST_PASS to run the live Kerberos tests")
	}
	return
}

// TestLiveKerberos_U2UAndPACDecode performs a U2U TGS exchange targeting the
// account itself, then decrypts the issued ticket with the client's TGT session
// key and decodes the embedded PAC.
func TestLiveKerberos_U2UAndPACDecode(t *testing.T) {
	kdc, realm, user, pass := liveBaseline(t)
	c := NewClient(user, realm, kdc).WithPassword(pass)
	if err := c.GetTGT(); err != nil {
		t.Fatalf("GetTGT: %v", err)
	}

	// Present the client's own TGT as the additional ticket; the KDC issues a
	// service ticket to the account encrypted under the TGT session key.
	ticket, ticketRaw, key, err := c.GetTGSU2U(user, "", c.tgtTicketRaw)
	if err != nil {
		t.Fatalf("GetTGSU2U(self): %v", err)
	}
	if len(ticketRaw) == 0 || len(key) == 0 {
		t.Fatal("GetTGSU2U(self) returned an empty ticket/key")
	}
	t.Logf("[ok] U2U ticket to self: sname=%v", ticket.SName.NameString)

	// The ticket enc-part is encrypted under the TGT session key (usage 2).
	plain, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageKDCRepTicket, ticket.EncPart.Cipher)
	if err != nil {
		t.Fatalf("decrypt U2U ticket enc-part with TGT session key: %v", err)
	}
	var enc messages.EncTicketPart
	if _, err := enc.Unmarshal(plain); err != nil {
		t.Fatalf("parse EncTicketPart: %v", err)
	}

	pacBytes := extractPACBytes(t, enc.AuthorizationData)
	if pacBytes == nil {
		t.Fatal("no AD-WIN2K-PAC found in the U2U ticket authorization data")
	}

	p, err := pac.Parse(pacBytes)
	if err != nil {
		t.Fatalf("pac.Parse: %v", err)
	}
	info, err := p.LogonInfo()
	if err != nil {
		t.Fatalf("pac.LogonInfo: %v", err)
	}
	if info.UserName() == "" {
		t.Fatal("decoded PAC logon-info has an empty EffectiveName")
	}
	t.Logf("[ok] decoded live PAC: user=%q rid=%d groups=%d", info.UserName(), info.UserRID(), info.GroupCount)
}

// extractPACBytes walks the AD-IF-RELEVANT wrapper for the AD-WIN2K-PAC element.
func extractPACBytes(t *testing.T, ad []messages.AuthorizationData) []byte {
	t.Helper()
	for _, e := range ad {
		if e.ADType != adTypeIfRelevant {
			continue
		}
		var inner []messages.AuthorizationData
		if _, err := asn1.Unmarshal(e.ADData, &inner); err != nil {
			t.Fatalf("unwrap AD-IF-RELEVANT: %v", err)
		}
		for _, ie := range inner {
			if ie.ADType == adTypeWin2KPAC {
				return ie.ADData
			}
		}
	}
	return nil
}
