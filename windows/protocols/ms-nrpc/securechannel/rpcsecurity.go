package securechannel

// IDL source: [MS-NRPC] — verified against the protocol's authoritative IDL
// (https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7).

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
)

// NetlogonSecurityContext adapts the Netlogon per-message MessageSecurity to the DCE/RPC
// client's SecurityContext, so a connection can protect its PDUs with RPC_C_AUTHN_NETLOGON
// (auth_type 0x44). Unlike NTLM, Netlogon signs and seals only the stub, not the whole PDU
// ([MS-NRPC] 3.3.4.2.1, confirmed against Windows: the token covers pduData, not the RPC
// header or sec_trailer), so both directions ignore the client's signedRegion argument and
// operate on the stub. A context is stateful (its sequence number advances per request) and
// must not be shared across connections.
type NetlogonSecurityContext struct {
	ms  *MessageSecurity
	aes bool
}

// NewNetlogonSecurityContext builds an RPC security provider from a secure channel: the
// cipher suite (AES vs legacy RC4) follows the channel, and the session key is taken from it.
// It is passed to the client's SetAuthProvider together with the NL_AUTH_MESSAGE bind token.
func NewNetlogonSecurityContext(sc *SecureChannel) *NetlogonSecurityContext {
	key := sc.SessionKey()
	if sc.UsesAES() {
		return &NetlogonSecurityContext{ms: NewMessageSecurityAES(key), aes: true}
	}
	return &NetlogonSecurityContext{ms: NewMessageSecurityRC4(key), aes: false}
}

// AuthValueLen is the token length: AES (NL_AUTH_SHA2_SIGNATURE) 56 sealing / 48 signing;
// legacy (NL_AUTH_SIGNATURE) 32 sealing / 24 signing. The Netlogon token is a fixed size, so
// the stub length is not consulted.
func (n *NetlogonSecurityContext) AuthValueLen(seal bool, _ int) int {
	if n.aes {
		if seal {
			return 56
		}
		return 48
	}
	if seal {
		return 32
	}
	return 24
}

// ProtectRequest seals (privacy) or signs (integrity) the request stub, returning the
// on-wire stub and the token.
func (n *NetlogonSecurityContext) ProtectRequest(_, stub []byte, seal bool) ([]byte, []byte, error) {
	if seal {
		return n.ms.Seal(stub)
	}
	token, err := n.ms.Sign(stub)
	return stub, token, err
}

// UnprotectResponse unseals (privacy) or verifies (integrity) the response stub against its
// token, returning the recovered plaintext stub.
func (n *NetlogonSecurityContext) UnprotectResponse(_, stub, authValue []byte, seal bool) ([]byte, error) {
	if seal {
		return n.ms.Unseal(stub, authValue)
	}
	return stub, n.ms.VerifySignature(stub, authValue)
}

// Compile-time check that the adapter satisfies the client's provider interface.
var _ client.SecurityContext = (*NetlogonSecurityContext)(nil)
