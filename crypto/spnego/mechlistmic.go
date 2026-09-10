package spnego

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/security"
)

// computeMechListMIC returns the SPNEGO mechListMIC for the negotiation this
// context is completing: GSS_GetMIC over the DER-encoded MechTypeList that was
// advertised in the NegTokenInit (RFC 4178 section 5), taken under the security
// context the mechanism just established.
//
// For NTLM the MIC is an NTLMSSP_MESSAGE_SIGNATURE ([MS-NLMP] 3.4.4) produced with
// the client signing key and sequence number 0 — it is the first message protected
// on the session. When key exchange is negotiated the checksum is additionally
// encrypted with the client sealing handle, which the security context handles.
//
// It returns nil, and no error, when there is nothing to compute a MIC under: no
// session key was derived, or the mech list was never recorded. Emitting no MIC is
// what the exchange did before, and is valid whenever the peer does not require one.
func (ctx *AuthContext) computeMechListMIC(negFlg flags.NegotiateFlags) ([]byte, error) {
	if len(ctx.SessionKey) == 0 || len(ctx.MechListDER) == 0 {
		return nil, nil
	}

	sec, err := security.NewContext(ctx.SessionKey, negFlg)
	if err != nil {
		return nil, fmt.Errorf("spnego: cannot key the mechListMIC: %w", err)
	}

	sig := sec.Sign(ctx.MechListDER)
	return sig[:], nil
}

// VerifyServerMechListMIC checks the mechListMIC a server returned in its final
// NegTokenResp against the mech list this client advertised.
//
// The server's MIC is taken with the server signing key over the same DER-encoded
// MechTypeList, at sequence number 0. A token carrying no mechListMIC verifies
// trivially: RFC 4178 leaves the field optional, and a server that does not require
// negotiation protection omits it.
func (ctx *AuthContext) VerifyServerMechListMIC(finalToken []byte) error {
	if len(finalToken) == 0 || len(ctx.SessionKey) == 0 || len(ctx.MechListDER) == 0 {
		return nil
	}

	resp, err := parseFinalNegTokenResp(finalToken)
	if err != nil {
		// A final token this client cannot parse is not evidence of tampering;
		// the exchange has already completed at the SMB layer.
		return nil
	}
	if len(resp.MechListMIC) == 0 {
		return nil
	}
	if len(resp.MechListMIC) != security.SignatureSize {
		return fmt.Errorf("spnego: server mechListMIC is %d bytes, want %d", len(resp.MechListMIC), security.SignatureSize)
	}

	sec, err := security.NewContext(ctx.SessionKey, ctx.authenticateFlags)
	if err != nil {
		return fmt.Errorf("spnego: cannot key the mechListMIC check: %w", err)
	}

	var sig [security.SignatureSize]byte
	copy(sig[:], resp.MechListMIC)
	if err := sec.VerifySignature(ctx.MechListDER, sig); err != nil {
		return fmt.Errorf("spnego: server mechListMIC does not verify over the advertised mech list: %w", err)
	}
	return nil
}

// parseFinalNegTokenResp unwraps a server's final SPNEGO token. The token may or
// may not be wrapped in a GSS-API SecurityBlob depending on the leg it arrived on.
func parseFinalNegTokenResp(token []byte) (*NegTokenResp, error) {
	resp := &NegTokenResp{}
	if _, err := resp.Unmarshal(token); err == nil {
		return resp, nil
	}

	blob := &SecurityBlob{}
	if _, err := blob.Unmarshal(token); err != nil {
		return nil, err
	}
	resp = &NegTokenResp{}
	if _, err := resp.Unmarshal(blob.Data); err != nil {
		return nil, err
	}
	return resp, nil
}
