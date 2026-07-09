package spnego

import (
	"errors"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
)

// CreateNegotiateToken creates the initial SPNEGO token with NTLM negotiate message
// Parameters:
//   - ctx: The authentication context containing domain, username, password, and other settings
//
// Returns:
//   - []byte: The SPNEGO token containing the NTLM negotiate message
//   - error: An error if token creation fails
func (ctx *AuthContext) CreateNegotiateToken(negotiateFlags flags.NegotiateFlags, version *version.Version) ([]byte, error) {
	switch ctx.Type {
	case AuthTypeNTLM:
		return ctx.processNegotiateInnerTokenNTLM(negotiateFlags, version)
	case AuthTypeKerberos:
		return ctx.processNegotiateInnerTokenKerberos(negotiateFlags, version)
	default:
		return nil, fmt.Errorf("unsupported authentication type: %v", ctx.Type)
	}
}

func (ctx *AuthContext) processNegotiateInnerTokenNTLM(negotiateFlags flags.NegotiateFlags, version *version.Version) ([]byte, error) {
	// Create NTLM NEGOTIATE message
	ntlmNegotiate, err := negotiate.CreateNegotiateMessage(ctx.Domain, ctx.Workstation, negotiateFlags, version)
	if err != nil {
		return nil, fmt.Errorf("failed to create NTLM NEGOTIATE message: %v", err)
	}

	ntlmNegotiateBytes, err := ntlmNegotiate.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NTLM NEGOTIATE message: %v", err)
	}

	// Retain the raw NEGOTIATE message for the AUTHENTICATE MIC computation.
	ctx.NegotiateMessageBytes = ntlmNegotiateBytes

	// Wrap in SPNEGO
	return CreateNegTokenInit(ntlmNegotiateBytes)
}

func (ctx *AuthContext) processNegotiateInnerTokenKerberos(_ flags.NegotiateFlags, _ *version.Version) ([]byte, error) {
	if ctx.Kerberos == nil {
		return nil, errors.New("spnego: kerberos provider not configured on AuthContext")
	}
	mechToken, err := ctx.Kerberos.InitToken()
	if err != nil {
		return nil, fmt.Errorf("spnego: kerberos init token: %w", err)
	}
	return CreateNegTokenInitKerberos(mechToken)
}
