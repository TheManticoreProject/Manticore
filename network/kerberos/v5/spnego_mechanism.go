package kerberos

import (
	"crypto/rand"
	"fmt"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// SPNEGOMechanism adapts the native Kerberos client and GSS-API layer to the
// crypto/spnego KerberosProvider interface, so SMB and DCE/RPC can authenticate
// with Kerberos through SPNEGO. It targets a single service principal (spn), for
// example "cifs/host.domain" for SMB or "host/host.domain" / "ldap/host.domain".
//
// It is created by the consumer (SMB/RPC) and assigned to
// spnego.AuthContext.Kerberos, keeping crypto/spnego free of any dependency on
// the Kerberos implementation.
type SPNEGOMechanism struct {
	client *KerberosClient
	spn    string
	ctx    *gssapi.SecContext
}

// NewSPNEGOMechanism builds a mechanism over an existing client (which supplies
// the credentials and realm) targeting the given service principal name.
func NewSPNEGOMechanism(client *KerberosClient, spn string) *SPNEGOMechanism {
	return &SPNEGOMechanism{client: client, spn: spn}
}

// InitToken acquires (if needed) a TGT and a service ticket for the SPN, then
// builds the KRB_AP_REQ GSS token to place in the SPNEGO NegTokenInit. Mutual
// authentication is requested and an initiator subkey is asserted (as Windows
// GSS clients do).
func (m *SPNEGOMechanism) InitToken() ([]byte, error) {
	// A preloaded (forged silver / captured) service ticket for the SPN needs no
	// TGT; otherwise acquire one if we do not already hold it.
	if !m.client.hasTGT && !m.client.hasPreloadedServiceTicket(m.spn) {
		if err := m.client.GetTGT(); err != nil {
			return nil, fmt.Errorf("kerberos spnego: GetTGT: %w", err)
		}
	}
	_, ticketRaw, key, keyEType, err := m.client.GetTGS(m.spn, true)
	if err != nil {
		return nil, fmt.Errorf("kerberos spnego: GetTGS %q: %w", m.spn, err)
	}
	subKey := make([]byte, kerbcrypto.KeyLen(keyEType))
	if _, err := rand.Read(subKey); err != nil {
		return nil, err
	}
	tok, ctx, err := gssapi.InitSecContext(gssapi.InitOptions{
		TicketRaw:    ticketRaw,
		SessionKey:   key,
		SessionEType: keyEType,
		ClientName:   messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{m.client.Username()}},
		ClientRealm:  m.client.Realm(),
		Flags:        gssapi.GSSMutualFlag | gssapi.GSSSequenceFlag | gssapi.GSSConfFlag | gssapi.GSSIntegFlag,
		Mutual:       true,
		SubKey:       subKey,
		SubKeyEType:  keyEType,
	})
	if err != nil {
		return nil, fmt.Errorf("kerberos spnego: InitSecContext: %w", err)
	}
	m.ctx = ctx
	return tok, nil
}

// AcceptResponseToken verifies the server's KRB_AP_REP (mutual authentication).
// An empty token is a no-op. A server KRB-ERROR is surfaced with its code.
func (m *SPNEGOMechanism) AcceptResponseToken(token []byte) error {
	if len(token) == 0 {
		return nil
	}
	if m.ctx == nil {
		return fmt.Errorf("kerberos spnego: no established context")
	}
	if tokID, krbMsg, err := gssapi.UnwrapToken(token); err == nil && tokID == gssapi.TokIDError {
		var ke messages.KRBError
		if _, e := ke.Unmarshal(krbMsg); e == nil {
			return fmt.Errorf("kerberos spnego: server returned KRB-ERROR %d: %s", ke.ErrorCode, ke.EText)
		}
	}
	return m.ctx.AcceptAPRep(token)
}

// SessionKey returns the established GSS context key for SMB/RPC message
// signing and sealing: the negotiated subkey if present, otherwise the service
// ticket session key.
func (m *SPNEGOMechanism) SessionKey() []byte {
	if m.ctx == nil {
		return nil
	}
	if len(m.ctx.SubKey) > 0 {
		return m.ctx.SubKey
	}
	return m.ctx.SessionKey
}
