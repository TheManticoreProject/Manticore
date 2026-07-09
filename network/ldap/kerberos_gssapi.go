package ldap

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/credentials"

	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// SASL GSSAPI security-layer bitmask (RFC 4752 §3.1): the server advertises the
// OR of the layers it supports and the client selects exactly one bit.
const (
	saslLayerNone            byte = 0x01 // no security layer (auth only)
	saslLayerIntegrity       byte = 0x02 // integrity: GSS_Wrap conf_flag = FALSE (sign)
	saslLayerConfidentiality byte = 0x04 // confidentiality: GSS_Wrap conf_flag = TRUE (seal)
)

// saslClientMaxRecv is the maximum SASL buffer size the client advertises it can
// receive (three octets, RFC 4752 §3.1). It matches the value Windows LDAP clients
// send and comfortably holds large search responses.
const saslClientMaxRecv = 0x00A00000

// nativeGSSAPIClient implements go-ldap's ldap.GSSAPIClient interface using
// Manticore's native, standard-library-only Kerberos + GSS-API stack, replacing
// the previous gokrb5 dependency. It drives the RFC 4752 SASL GSSAPI handshake:
// an AP-REQ (with mutual auth), the AP-REP, then the security-layer negotiation,
// after which it can sign and/or seal subsequent LDAP PDUs.
type nativeGSSAPIClient struct {
	kc    *kerberos.KerberosClient
	realm string
	user  string
	ctx   *gssapi.SecContext

	// desiredLayer is the SASL security layer the client wants to negotiate
	// (defaults to no security layer, preserving auth-only behaviour).
	desiredLayer byte
	// channelBindings is the marshalled GSS-API channel-bindings buffer to hash
	// into the AP-REQ authenticator checksum (nil = GSS_C_NO_BINDINGS). It is set
	// for LDAPS binds to carry the tls-server-end-point token.
	channelBindings []byte

	// negotiatedLayer and serverMaxSend record the outcome of NegotiateSaslAuth so
	// the session can install the matching security layer on the connection.
	negotiatedLayer byte
	serverMaxSend   int
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
	return &nativeGSSAPIClient{kc: kc, realm: strings.ToUpper(realm), user: user, desiredLayer: saslLayerNone}, nil
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
		_, ticketRaw, key, keyEType, err := g.kc.GetTGS(target, true)
		if err != nil {
			return nil, false, fmt.Errorf("ldap gssapi: GetTGS %q: %w", target, err)
		}
		// Assert an initiator subkey of the session-key etype, as Windows GSS
		// clients do; AD's RFC 4121 (AES) acceptor expects one.
		subKey := make([]byte, kerbcrypto.KeyLen(keyEType))
		if _, err := rand.Read(subKey); err != nil {
			return nil, false, err
		}
		tok, ctx, err := gssapi.InitSecContext(gssapi.InitOptions{
			TicketRaw:       ticketRaw,
			SessionKey:      key,
			SessionEType:    keyEType,
			ClientName:      messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{g.user}},
			ClientRealm:     g.realm,
			Flags:           gssapi.GSSMutualFlag | gssapi.GSSSequenceFlag | gssapi.GSSConfFlag | gssapi.GSSIntegFlag,
			ChannelBindings: g.channelBindings,
			Mutual:          true,
			SubKey:          subKey,
			SubKeyEType:     keyEType,
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
// max receive buffer), selects the client's desired layer among those offered,
// and returns a GSS-wrapped (integrity-only, per RFC 4752) response carrying the
// selected layer, the client's max receive buffer, and the optional authzid. The
// negotiated layer and the server's max buffer are recorded so the session can
// wrap subsequent PDUs accordingly.
func (g *nativeGSSAPIClient) NegotiateSaslAuth(token []byte, authzid string) ([]byte, error) {
	offer, _, err := g.ctx.Unwrap(token)
	if err != nil {
		return nil, fmt.Errorf("ldap gssapi: unwrap SASL offer: %w", err)
	}

	resp, chosen, serverMax, err := selectSASLLayer(offer, g.desiredLayer, authzid)
	if err != nil {
		return nil, err
	}

	// The RFC 4752 §3.1 response is always integrity-only (conf_flag = FALSE),
	// regardless of the layer being selected.
	out, err := g.ctx.Wrap(resp, false)
	if err != nil {
		return nil, fmt.Errorf("ldap gssapi: wrap SASL response: %w", err)
	}
	g.negotiatedLayer = chosen
	g.serverMaxSend = serverMax
	return out, nil
}

// selectSASLLayer parses the RFC 4752 §3.1 server security-layer offer (first
// octet = supported-layer bitmask, next three = the server's max receive buffer
// in network byte order), picks the desired layer among those offered, and builds
// the unwrapped client response (selected-layer octet, the client's own max
// receive buffer, then the authzid). It returns the response, the chosen layer,
// and the server's max buffer size. It is the pure, GSS-free core of the
// negotiation so it can be unit-tested without a security context.
func selectSASLLayer(offer []byte, desired byte, authzid string) (resp []byte, chosen byte, serverMax int, err error) {
	if len(offer) < 4 {
		return nil, 0, 0, fmt.Errorf("ldap gssapi: short SASL offer (%d bytes)", len(offer))
	}
	supported := offer[0]
	serverMax = int(offer[1])<<16 | int(offer[2])<<8 | int(offer[3])

	// The server always offers "no security layer"; default to it if unset.
	chosen = desired
	if chosen == 0 {
		chosen = saslLayerNone
	}
	if supported&chosen == 0 {
		return nil, 0, 0, fmt.Errorf("ldap gssapi: server does not offer requested security layer (offered 0x%02x, wanted 0x%02x)", supported, chosen)
	}

	resp = make([]byte, 4+len(authzid))
	resp[0] = chosen
	if chosen != saslLayerNone {
		// Advertise the client's maximum receive buffer (three octets, network
		// byte order); RFC 4752 requires it to be 0 when no layer is selected.
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(saslClientMaxRecv))
		copy(resp[1:4], b[1:4])
	}
	copy(resp[4:], authzid)
	return resp, chosen, serverMax, nil
}

// securityLayer returns the saslCipher for the negotiated layer, or nil if the
// bind negotiated no security layer (in which case the connection stays a
// pass-through). It must be called after NegotiateSaslAuth.
func (g *nativeGSSAPIClient) securityLayer() saslCipher {
	if g.negotiatedLayer == 0 || g.negotiatedLayer == saslLayerNone {
		return nil
	}
	return &gssSASLCipher{
		ctx:  g.ctx,
		seal: g.negotiatedLayer == saslLayerConfidentiality,
		max:  g.serverMaxSend,
	}
}

// DeleteSecContext releases the TGT key material held by the client. The
// per-message security context (service-ticket / acceptor-subkey material) is
// retained so an established SASL security layer keeps working after the bind.
func (g *nativeGSSAPIClient) DeleteSecContext() error {
	g.kc.Destroy()
	return nil
}
