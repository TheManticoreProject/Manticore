package kerberos

import (
	"fmt"
	"net"
	"strings"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// fakeS4UKDC serves TGS-REQs on a loopback UDP socket, answering each with a
// TGS-REP for the requested sname whose cname is chosen by replyClient and whose
// enc-part is sealed under the requesting client's TGT session key. If ticketFor
// is non-nil it supplies the issued ticket (to carry a decryptable enc-part).
func fakeS4UKDC(t *testing.T, c *KerberosClient, replyClient func(n int) string, ticketFor func(sname messages.PrincipalName) messages.Ticket) {
	t.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	c.kdcHost = conn.LocalAddr().String()

	go func() {
		buf := make([]byte, 65535)
		for n := 0; ; n++ {
			size, peer, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			var req messages.TGSReq
			if _, err := req.Unmarshal(buf[:size]); err != nil {
				return
			}
			ticket := messages.Ticket{
				TktVno:  messages.KerberosV5,
				EncPart: messages.EncryptedData{EType: messages.ETypeAES256CTSHMACSHA196, Cipher: []byte(fmt.Sprintf("ticket-%d-opaque-cipher-bytes", n))},
			}
			if ticketFor != nil {
				ticket = ticketFor(req.ReqBody.SName)
			}
			ticket.Realm = c.realm
			ticket.SName = req.ReqBody.SName
			enc := messages.EncTGSRepPart{
				Key:    messages.EncryptionKey{KeyType: messages.ETypeAES256CTSHMACSHA196, KeyValue: make([]byte, 32)},
				Nonce:  req.ReqBody.Nonce,
				Flags:  messages.NewKerberosFlags(messages.TicketFlagForwardable),
				SRealm: c.realm,
				SName:  req.ReqBody.SName,
			}
			encBytes, err := enc.Marshal()
			if err != nil {
				return
			}
			cipher, err := kerbcrypto.Encrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSRepEncSessionKey, encBytes)
			if err != nil {
				return
			}
			rep := messages.TGSRep{
				CRealm:  c.realm,
				CName:   messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{replyClient(n)}},
				Ticket:  ticket,
				EncPart: messages.EncryptedData{EType: c.sessionEType, Cipher: cipher},
			}
			repBytes, err := rep.Marshal()
			if err != nil {
				return
			}
			if _, err := conn.WriteToUDP(repBytes, peer); err != nil {
				return
			}
		}
	}()
}

// TestS4U2SelfAndProxyAcceptImpersonatedClient drives S4U2Self then S4U2Proxy
// against a fake KDC that, as [MS-SFU] specifies, names the impersonated user
// (not the requesting service) as the reply cname. Both exchanges must succeed.
func TestS4U2SelfAndProxyAcceptImpersonatedClient(t *testing.T) {
	c := fakeTGTClient(t) // service "alice"
	fakeS4UKDC(t, c, func(int) string { return "Administrator" }, nil)

	_, selfRaw, _, _, err := c.S4U2Self("Administrator", "")
	if err != nil {
		t.Fatalf("S4U2Self rejected a reply naming the impersonated user: %v", err)
	}
	if _, _, _, _, err := c.S4U2Proxy("cifs/fs01.corp.local", selfRaw); err != nil {
		t.Fatalf("S4U2Proxy rejected a reply naming the evidence ticket's client: %v", err)
	}
}

// TestS4U2SelfRejectsUnexpectedClient checks the client binding is still
// enforced: an S4U2Self reply naming someone other than the impersonated user
// is rejected.
func TestS4U2SelfRejectsUnexpectedClient(t *testing.T) {
	c := fakeTGTClient(t)
	fakeS4UKDC(t, c, func(int) string { return "mallory" }, nil)

	_, _, _, _, err := c.S4U2Self("Administrator", "")
	if err == nil || !strings.Contains(err.Error(), "client identity") {
		t.Fatalf("S4U2Self error = %v, want client identity mismatch", err)
	}
}

// TestSapphireHarvestAcceptsImpersonatedClient checks the sapphire
// S4U2Self+U2U harvest accepts a reply naming the impersonated user.
func TestSapphireHarvestAcceptsImpersonatedClient(t *testing.T) {
	c := fakeTGTClient(t)
	pacBytes := []byte("genuine-pac-bytes")
	harvested := syntheticTicketWithPAC(t, pacBytes, "Administrator", c.realm, c.sessionKey, c.sessionEType)
	fakeS4UKDC(t, c, func(int) string { return "Administrator" }, func(messages.PrincipalName) messages.Ticket { return harvested })

	_, cname, _, err := c.harvestPACViaS4USelfU2U("Administrator", "")
	if err != nil {
		t.Fatalf("sapphire harvest rejected a reply naming the impersonated user: %v", err)
	}
	if len(cname.NameString) != 1 || cname.NameString[0] != "Administrator" {
		t.Fatalf("harvested client = %+v, want Administrator", cname)
	}
}
