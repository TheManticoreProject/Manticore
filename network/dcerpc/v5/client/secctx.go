package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/security"
)

// securityContext protects outbound request stubs and unprotects inbound response stubs for
// an authenticated connection-oriented RPC session. Each security provider (NTLM today,
// Netlogon next) supplies its own auth_value token format and chooses which bytes it signs:
// NTLM signs the whole PDU and so uses signedRegion, whereas a provider that signs only the
// stub can ignore signedRegion and operate on stub. The client owns PDU framing (stub
// padding, header fields, sec_trailer), which is provider-independent ([MS-RPCE] 2.2.2.11).
type securityContext interface {
	// AuthValueLen returns the auth_value byte length this provider emits, given whether the
	// request is sealed (PKT_PRIVACY) or only signed (PKT/PKT_INTEGRITY). It must be known
	// before the PDU header is marshalled, since it feeds auth_length and frag_length.
	AuthValueLen(seal bool) int

	// ProtectRequest produces the on-wire stub and the auth_value for one outbound request
	// fragment. signedRegion is the marshalled PDU from the header through the sec_trailer
	// (auth_value excluded) computed over the plaintext stub; stub is the padded stub as a
	// separate buffer. When seal is true the returned stub is encrypted, otherwise it equals
	// stub. signedRegion MUST NOT be mutated.
	ProtectRequest(signedRegion, stub []byte, seal bool) (onWireStub, authValue []byte, err error)

	// UnprotectResponse recovers the plaintext stub from one inbound response fragment.
	// signedRegion and stub are overlapping views into the same fragment buffer — stub is a
	// sub-slice of signedRegion — so when seal is true the provider decrypts the stub in
	// place (which updates signedRegion) before verifying integrity. It returns the recovered
	// plaintext stub (still including any auth padding, which the caller strips).
	UnprotectResponse(signedRegion, stub, authValue []byte, seal bool) (plainStub []byte, err error)
}

// ntlmSecurityContext adapts an NTLM *security.Context to securityContext. NTLM's token is a
// fixed 16-byte MESSAGE_SIGNATURE and its signature covers the whole PDU, so both directions
// operate over signedRegion; only the stub is sealed for PKT_PRIVACY ([MS-RPCE] 3.3, NTLM2).
type ntlmSecurityContext struct{ ctx *security.Context }

// AuthValueLen is the fixed NTLM signature size regardless of seal level.
func (n *ntlmSecurityContext) AuthValueLen(bool) int { return security.SignatureSize }

// ProtectRequest signs signedRegion (and, when sealing, encrypts the stub) exactly as the
// NTLM2 sender rules require.
func (n *ntlmSecurityContext) ProtectRequest(signedRegion, stub []byte, seal bool) ([]byte, []byte, error) {
	if seal {
		onWire, sig := n.ctx.SealWith(signedRegion, stub)
		return onWire, sig[:], nil
	}
	sig := n.ctx.Sign(signedRegion)
	return stub, sig[:], nil
}

// UnprotectResponse decrypts the stub in place (when sealed) and then verifies the signature
// over the whole PDU, matching the NTLM2 receiver rules.
func (n *ntlmSecurityContext) UnprotectResponse(signedRegion, stub, authValue []byte, seal bool) ([]byte, error) {
	if len(authValue) != security.SignatureSize {
		return nil, fmt.Errorf("unexpected auth_length %d, want %d", len(authValue), security.SignatureSize)
	}
	var sig [security.SignatureSize]byte
	copy(sig[:], authValue)
	if seal {
		n.ctx.DecryptInbound(stub)
	}
	if err := n.ctx.VerifySignature(signedRegion, sig); err != nil {
		return nil, err
	}
	return stub, nil
}
