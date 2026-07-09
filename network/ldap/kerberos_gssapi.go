package ldap

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/credentials"

	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// nativeGSSAPIClient implements go-ldap's ldap.GSSAPIClient interface using
// Manticore's native, standard-library-only Kerberos + GSS-API stack, replacing
// the previous gokrb5 dependency. It drives the RFC 4752 SASL GSSAPI handshake:
// an AP-REQ (with mutual auth), the AP-REP, then the security-layer negotiation.
type nativeGSSAPIClient struct {
	kc    *kerberos.KerberosClient
	realm string
	user  string
	ctx   *gssapi.SecContext
}

// newNativeGSSAPIClient builds a GSSAPI client and acquires a TGT for the given
// credentials (password, or NT hash for overpass-the-hash). kdc is the KDC host
// (the DC), realm the Kerberos realm.
func newNativeGSSAPIClient(kdc, realm string, creds *credentials.Credentials) (*nativeGSSAPIClient, error) {
	user := creds.GetUsername()
	kc := kerberos.NewClient(user, realm, kdc)
	if creds.CanPassTheHash() {
		if err := kc.WithNTHash(creds.GetNTHash()); err != nil {
			return nil, fmt.Errorf("ldap gssapi: NT hash: %w", err)
		}
	} else {
		kc.WithPassword(creds.GetPassword())
	}
	if err := kc.GetTGT(); err != nil {
		return nil, fmt.Errorf("ldap gssapi: GetTGT: %w", err)
	}
	return &nativeGSSAPIClient{kc: kc, realm: strings.ToUpper(realm), user: user}, nil
}

// InitSecContext establishes the GSS security context. On the first call
// (token == nil) it acquires a service ticket for the target SPN and returns the
// KRB_AP_REQ token, requesting a reply (needContinue = true). On the second call
// it verifies the KRB_AP_REP for mutual authentication.
func (g *nativeGSSAPIClient) InitSecContext(target string, token []byte) ([]byte, bool, error) {
	return g.InitSecContextWithOptions(target, token, nil)
}

// InitSecContextWithOptions is InitSecContext with go-ldap's options parameter,
// which this native mechanism does not need.
func (g *nativeGSSAPIClient) InitSecContextWithOptions(target string, token []byte, _ []int) ([]byte, bool, error) {
	if g.ctx == nil {
		ticket, ticketRaw, key, err := g.kc.GetTGS(target, true)
		if err != nil {
			return nil, false, fmt.Errorf("ldap gssapi: GetTGS %q: %w", target, err)
		}
		// Assert an initiator subkey of the session-key etype, as Windows GSS
		// clients do; AD's RFC 4121 (AES) acceptor expects one.
		subKey := make([]byte, kerbcrypto.KeyLen(ticket.EncPart.EType))
		if _, err := rand.Read(subKey); err != nil {
			return nil, false, err
		}
		tok, ctx, err := gssapi.InitSecContext(gssapi.InitOptions{
			TicketRaw:    ticketRaw,
			SessionKey:   key,
			SessionEType: ticket.EncPart.EType,
			ClientName:   messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{g.user}},
			ClientRealm:  g.realm,
			Flags:        gssapi.GSSMutualFlag | gssapi.GSSSequenceFlag | gssapi.GSSConfFlag | gssapi.GSSIntegFlag,
			Mutual:       true,
			SubKey:       subKey,
			SubKeyEType:  ticket.EncPart.EType,
		})
		if err != nil {
			return nil, false, fmt.Errorf("ldap gssapi: InitSecContext: %w", err)
		}
		g.ctx = ctx
		return tok, true, nil
	}

	// Second call: token should be the server's KRB_AP_REP. If the acceptor
	// rejected the AP-REQ it returns a wrapped KRB-ERROR instead; surface it.
	if tokID, krbMsg, err := gssapi.UnwrapToken(token); err == nil && tokID == gssapi.TokIDError {
		var ke messages.KRBError
		if _, e := ke.Unmarshal(krbMsg); e == nil {
			return nil, false, fmt.Errorf("ldap gssapi: server returned KRB-ERROR %d: %s", ke.ErrorCode, ke.EText)
		}
	}
	if err := g.ctx.AcceptAPRep(token); err != nil {
		return nil, false, fmt.Errorf("ldap gssapi: AP-REP: %w", err)
	}
	return nil, false, nil
}

// NegotiateSaslAuth performs the RFC 4752 §3.1 final step: it GSS-unwraps the
// server's security-layer offer (first octet = supported layers, next three =
// max receive buffer), selects "no security layer", and returns a GSS-wrapped
// (integrity-only) response carrying the selection and optional authzid.
func (g *nativeGSSAPIClient) NegotiateSaslAuth(token []byte, authzid string) ([]byte, error) {
	offer, _, err := g.ctx.Unwrap(token)
	if err != nil {
		return nil, fmt.Errorf("ldap gssapi: unwrap SASL offer: %w", err)
	}
	if len(offer) < 4 {
		return nil, fmt.Errorf("ldap gssapi: short SASL offer (%d bytes)", len(offer))
	}

	// Select "no security layer" (0x01) with a zero max buffer; append authzid.
	resp := make([]byte, 4+len(authzid))
	resp[0] = 0x01
	copy(resp[4:], authzid)

	out, err := g.ctx.Wrap(resp, false)
	if err != nil {
		return nil, fmt.Errorf("ldap gssapi: wrap SASL response: %w", err)
	}
	return out, nil
}

// DeleteSecContext releases the key material held by the client.
func (g *nativeGSSAPIClient) DeleteSecContext() error {
	g.kc.Destroy()
	return nil
}
