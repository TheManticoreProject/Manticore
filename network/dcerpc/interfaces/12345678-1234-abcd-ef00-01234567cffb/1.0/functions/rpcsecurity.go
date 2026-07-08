package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

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
	ms *MessageSecurity
}

// NewNetlogonSecurityContext builds an RPC security provider for the AES cipher suite from a
// secure-channel session key (SecureChannel.SessionKey). It is passed to the client's
// SetAuthProvider together with the NL_AUTH_MESSAGE bind token.
func NewNetlogonSecurityContext(sessionKey [16]byte) *NetlogonSecurityContext {
	return &NetlogonSecurityContext{ms: NewMessageSecurityAES(sessionKey)}
}

// AuthValueLen is the NL_AUTH_SHA2_SIGNATURE length: 56 octets when sealing (the confounder
// is present), 48 when only signing.
func (n *NetlogonSecurityContext) AuthValueLen(seal bool) int {
	if seal {
		return 56
	}
	return 48
}

// ProtectRequest seals (privacy) or signs (integrity) the request stub, returning the
// on-wire stub and the NL_AUTH_SHA2_SIGNATURE token.
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
