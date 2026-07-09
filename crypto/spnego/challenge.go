package spnego

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
)

// decodeNTHash decodes a hex-encoded NT hash (32 hex characters) into the 16-byte array
// the NTLMv2 pass-the-hash path expects.
func decodeNTHash(ntHashHex string) ([16]byte, error) {
	var ntHash [16]byte
	raw, err := hex.DecodeString(ntHashHex)
	if err != nil {
		return ntHash, err
	}
	if len(raw) != 16 {
		return ntHash, fmt.Errorf("NT hash must be 16 bytes (32 hex chars), got %d bytes", len(raw))
	}
	copy(ntHash[:], raw)
	return ntHash, nil
}

// CreateAuthenticateTokenFromChallengeToken processes the server's challenge token and creates an authenticate token
// Parameters:
//   - challengeToken: The SPNEGO token containing the server's challenge
//
// Returns:
//   - []byte: The SPNEGO token containing the authenticate message
//   - error: An error if token processing fails
func (ctx *AuthContext) CreateAuthenticateTokenFromChallengeToken(challengeToken []byte) ([]byte, error) {

	// First, unpack the security blob
	securityBlob := &SecurityBlob{}
	_, err := securityBlob.Unmarshal(challengeToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SPNEGO token: %v", err)
	}

	// Then, unpack the NegTokenResp
	resp := NegTokenResp{}
	_, err = resp.Unmarshal(securityBlob.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SPNEGO token: %v", err)
	}

	// Check if the server accepted our mechanism
	if resp.NegState == NegStateReject {
		return nil, errors.New("server rejected authentication")
	}

	if resp.SupportedMech.Equal(NtlmOID) {
		return ctx.processChallengeInnerTokenNTLM(resp.ResponseToken)
	} else if resp.SupportedMech.Equal(KerberosOID) {
		return ctx.processChallengeInnerTokenKerberos(resp.ResponseToken)
	} else {
		return nil, fmt.Errorf("unsupported authentication type: %v", resp.SupportedMech)
	}
}

// processChallengeInnerTokenNTLM processes the NTLM challenge token and creates an NTLM authenticate token
// Parameters:
//   - innerToken: The inner NTLM challenge token bytes
//
// Returns:
//   - []byte: The SPNEGO token containing the NTLM authenticate message
//   - error: An error if token processing fails
func (ctx *AuthContext) processChallengeInnerTokenNTLM(innerToken []byte) ([]byte, error) {
	// Parse the NTLM CHALLENGE message
	challenge := &challenge.ChallengeMessage{}
	_, err := challenge.Unmarshal(innerToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NTLM CHALLENGE message: %v", err)
	}

	// Store the challenge for later use
	ctx.NTLMChallenge = challenge

	// Create NTLM AUTHENTICATE message. When an NT hash is supplied (pass-the-hash),
	// compute the response from it instead of the password.
	var ntlmAuth *authenticate.AuthenticateMessage
	if ctx.NTHash != "" {
		ntHash, derr := decodeNTHash(ctx.NTHash)
		if derr != nil {
			return nil, fmt.Errorf("invalid NT hash for pass-the-hash: %v", derr)
		}
		ntlmAuth, err = authenticate.CreateAuthenticateMessageWithNTHash(challenge, ctx.Username, ntHash, ctx.Domain, ctx.Workstation)
	} else {
		ntlmAuth, err = authenticate.CreateAuthenticateMessage(challenge, ctx.Username, ctx.Password, ctx.Domain, ctx.Workstation)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create NTLM AUTHENTICATE message: %v", err)
	}

	// Retain the derived session key so callers can use it as the SMB signing MAC key.
	ctx.SessionKey = ntlmAuth.SessionKey

	// When the server's CHALLENGE carried an MsvAvTimestamp, CreateAuthenticateMessage
	// sets the MsvAvFlags MIC-present bit in the NtChallengeResponse, so the
	// AUTHENTICATE MUST also carry a matching MIC over NEGOTIATE||CHALLENGE||
	// AUTHENTICATE (MS-NLMP 3.1.5.1.2). innerToken is the raw CHALLENGE_MESSAGE.
	if ntlmAuth.NeedsMIC {
		if len(ctx.NegotiateMessageBytes) == 0 {
			return nil, fmt.Errorf("cannot compute NTLM MIC: NEGOTIATE message not retained")
		}
		if err := ntlmAuth.ComputeMIC(ctx.NegotiateMessageBytes, innerToken); err != nil {
			return nil, fmt.Errorf("failed to compute NTLM MIC: %v", err)
		}
	}

	ntlmAuthBytes, err := ntlmAuth.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NTLM AUTHENTICATE message: %v", err)
	}

	// The client's continuation (AUTHENTICATE) token carries only the responseToken
	// — no negState and no supportedMech — matching the Windows client.
	negTokenResp := NegTokenResp{SuppressNegState: true}
	negTokenResp.SetMechToken(ntlmAuthBytes)

	marshalledNegTokenResp, err := negTokenResp.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NegTokenResp: %v", err)
	}

	securityBlob := SecurityBlob{}
	securityBlob.Data = marshalledNegTokenResp

	marshalledSecurityBlob, err := securityBlob.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SecurityBlob: %v", err)
	}

	return marshalledSecurityBlob, nil
}

// processChallengeInnerTokenKerberos processes the Kerberos challenge token and creates a Kerberos authenticate token
// Parameters:
//   - innerToken: The inner Kerberos challenge token bytes
//
// Returns:
//   - []byte: The SPNEGO token containing the Kerberos authenticate message
//   - error: An error if token processing fails
func (ctx *AuthContext) processChallengeInnerTokenKerberos(innerToken []byte) ([]byte, error) {
	if ctx.Kerberos == nil {
		return nil, errors.New("spnego: kerberos provider not configured on AuthContext")
	}
	// innerToken is the server's response token: a KRB_AP_REP for mutual
	// authentication (empty if the server asserted none). Verifying it also lets
	// the provider adopt any acceptor subkey.
	if err := ctx.Kerberos.AcceptResponseToken(innerToken); err != nil {
		return nil, fmt.Errorf("spnego: kerberos accept response: %w", err)
	}
	// The GSS context is established once the AP-REP is verified; capture the
	// session key for message signing/sealing. Kerberos needs no further token.
	ctx.SessionKey = ctx.Kerberos.SessionKey()
	return nil, nil
}
