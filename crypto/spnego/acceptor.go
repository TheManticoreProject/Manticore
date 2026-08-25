package spnego

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv1"
	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/crypto/rc4"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/header"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/types"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// Verification outcomes reported by AcceptContext.Verify. They are distinguished
// because a caller acts differently on each: a capture-only acceptor expects
// ErrCaptureOnly on every exchange, whereas a serving acceptor treats an unknown
// identity and a wrong password the same way on the wire but wants them apart in
// its logs.
var (
	// ErrCaptureOnly reports that the AUTHENTICATE was recorded but not verified,
	// because no CredentialLookup is configured. The captured material is
	// available regardless; this is the expected outcome for an acceptor whose
	// purpose is to harvest responses rather than to serve.
	ErrCaptureOnly = errors.New("spnego: authenticate recorded but not verified, no credential lookup configured")

	// ErrNoAuthenticate reports that Verify was called before an AUTHENTICATE was
	// accepted.
	ErrNoAuthenticate = errors.New("spnego: no AUTHENTICATE message has been accepted")

	// ErrUnknownIdentity reports that CredentialLookup had no credential for the
	// identity the client claimed.
	ErrUnknownIdentity = errors.New("spnego: no credential for the claimed identity")

	// ErrBadResponse reports that the NT challenge response did not verify
	// against the credential, which is what a wrong password looks like.
	ErrBadResponse = errors.New("spnego: NT challenge response did not verify")

	// ErrBadMIC reports that the NT challenge response verified but the message
	// integrity code did not, meaning the exchange was tampered with in flight.
	ErrBadMIC = errors.New("spnego: AUTHENTICATE message integrity code did not verify")
)

// AcceptContext is the acceptor side of an NTLM authentication carried over
// SPNEGO: the counterpart of AuthContext, which initiates one.
//
// The exchange is two calls. AcceptNegotiateToken consumes the client's
// NEGOTIATE and returns the CHALLENGE to send back. AcceptAuthenticateToken
// consumes the client's AUTHENTICATE and records it. Verification is a separate
// step, so an acceptor that only wants to harvest responses never has to hold a
// credential, while one that serves calls Verify to decide.
//
// A context handles one exchange and is not safe for concurrent use.
type AcceptContext struct {
	// TargetName is the name advertised in the CHALLENGE. Leaving it empty omits
	// it, which a client is entitled to refuse.
	TargetName string

	// TargetType declares whether TargetName names a domain or a server.
	TargetType challenge.TargetType

	// TargetInfo is the AV_PAIR list advertised in the CHALLENGE, as built by
	// targetinfo.BuildServerTargetInfo. A client folds these bytes into its
	// NTLMv2 blob, so they are covered by the response and cannot be changed
	// after the CHALLENGE is sent.
	//
	// Supplying an MsvAvTimestamp here obliges the client to carry a MIC, which
	// Verify then checks; omit it if the acceptor would rather not require one.
	TargetInfo []byte

	// ServerChallenge is the 8-byte nonce. The zero value means generate one from
	// the system's cryptographic random source when the CHALLENGE is built, which
	// is what an acceptor should normally do: a predictable or reused challenge
	// lets a captured response be replayed.
	ServerChallenge [8]byte

	// Version is advertised in the CHALLENGE. Nil omits it.
	Version *version.Version

	// CredentialLookup resolves a claimed identity to the account's NT hash.
	// Returning false means the identity is unknown.
	//
	// Nil makes the acceptor capture-only: Verify records nothing further and
	// reports ErrCaptureOnly. Holding NT hashes rather than passwords is
	// deliberate — the hash is all that verification needs.
	CredentialLookup func(domain, username string) (ntHash [16]byte, ok bool)

	// Negotiate and NegotiateMessageBytes hold the client's NEGOTIATE, the
	// latter because the MIC is computed over the raw message.
	Negotiate             *negotiate.NegotiateMessage
	NegotiateMessageBytes []byte

	// Challenge and ChallengeMessageBytes hold the CHALLENGE that was issued,
	// for the same reason.
	Challenge             *challenge.ChallengeMessage
	ChallengeMessageBytes []byte

	// Authenticate and AuthenticateMessageBytes hold the client's AUTHENTICATE.
	Authenticate             *authenticate.AuthenticateMessage
	AuthenticateMessageBytes []byte

	// SessionKey is the ExportedSessionKey, set by a successful Verify. It is the
	// key a consumer uses for message signing and sealing. It stays nil until
	// verification succeeds, because deriving it requires the credential.
	SessionKey []byte

	// Verified records whether Verify has succeeded.
	Verified bool
}

// NewAcceptContext creates an acceptor advertising the given identity.
//
// Parameters:
//   - targetName: the name to advertise in the CHALLENGE
//   - targetType: whether targetName names a domain or a server
//   - targetInfo: the AV_PAIR list to advertise, or nil
//
// Returns:
//   - The acceptor context
func NewAcceptContext(targetName string, targetType challenge.TargetType, targetInfo []byte) *AcceptContext {
	return &AcceptContext{
		TargetName: targetName,
		TargetType: targetType,
		TargetInfo: targetInfo,
	}
}

// AcceptNegotiateToken consumes the client's NEGOTIATE token and returns the
// SPNEGO token carrying the CHALLENGE to send back.
//
// The returned bytes are framed exactly as CreateAuthenticateTokenFromChallengeToken
// expects to receive them: a SecurityBlob wrapping a NegTokenResp whose negState
// is accept-incomplete and whose supportedMech is the NTLM OID.
//
// Parameters:
//   - token: the client's NEGOTIATE, as a SPNEGO token or a bare NTLMSSP message
//
// Returns:
//   - The SPNEGO token carrying the CHALLENGE
//   - An error if the client's token cannot be answered
func (ctx *AcceptContext) AcceptNegotiateToken(token []byte) ([]byte, error) {
	inner, err := extractNTLMMessage(token)
	if err != nil {
		return nil, fmt.Errorf("failed to extract the NTLM NEGOTIATE message: %v", err)
	}
	if err := expectMessageType(inner, types.MESSAGE_TYPE_NEGOTIATE); err != nil {
		return nil, err
	}

	negotiateMessage := &negotiate.NegotiateMessage{}
	if _, err := negotiateMessage.Unmarshal(inner); err != nil {
		return nil, fmt.Errorf("failed to parse the NTLM NEGOTIATE message: %v", err)
	}
	ctx.Negotiate = negotiateMessage
	// The MIC is computed over the raw message, so keep the bytes as received
	// rather than a re-marshalling of them.
	ctx.NegotiateMessageBytes = append([]byte(nil), inner...)

	// A zero challenge means one was not supplied, so generate one. Reusing a
	// challenge across exchanges would let a response captured from one be
	// replayed into another.
	if ctx.ServerChallenge == ([8]byte{}) {
		generated, err := challenge.NewServerChallenge()
		if err != nil {
			return nil, err
		}
		ctx.ServerChallenge = generated
	}

	challengeMessage, err := challenge.CreateChallengeMessage(
		negotiateMessage, ctx.ServerChallenge, ctx.TargetName, ctx.TargetType, ctx.TargetInfo, ctx.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to build the NTLM CHALLENGE message: %v", err)
	}

	challengeBytes, err := challengeMessage.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the NTLM CHALLENGE message: %v", err)
	}
	ctx.Challenge = challengeMessage
	ctx.ChallengeMessageBytes = challengeBytes

	// The CHALLENGE is a continuation token: a NegTokenResp naming the mechanism
	// and reporting that more messages are expected, wrapped in a SecurityBlob.
	negTokenResp := NewNegTokenResp(NegStateAcceptIncomplete, NtlmOID, challengeBytes)
	marshalledNegTokenResp, err := negTokenResp.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the NegTokenResp: %v", err)
	}

	securityBlob := SecurityBlob{Data: marshalledNegTokenResp}
	marshalledSecurityBlob, err := securityBlob.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the SecurityBlob: %v", err)
	}

	return marshalledSecurityBlob, nil
}

// AcceptAuthenticateToken consumes the client's AUTHENTICATE token and records
// it, without verifying anything.
//
// Recording and verifying are separate so that a failure to authenticate does not
// lose the response: the material is captured whether or not a credential exists
// to check it against. Call Verify afterwards to decide whether to honour it.
//
// Parameters:
//   - token: the client's AUTHENTICATE, as a SPNEGO token or a bare NTLMSSP message
//
// Returns:
//   - An error if the token cannot be parsed
func (ctx *AcceptContext) AcceptAuthenticateToken(token []byte) error {
	if ctx.Challenge == nil {
		return fmt.Errorf("cannot accept an AUTHENTICATE before a CHALLENGE has been issued")
	}

	inner, err := extractNTLMMessage(token)
	if err != nil {
		return fmt.Errorf("failed to extract the NTLM AUTHENTICATE message: %v", err)
	}
	if err := expectMessageType(inner, types.MESSAGE_TYPE_AUTHENTICATE); err != nil {
		return err
	}

	authenticateMessage := &authenticate.AuthenticateMessage{}
	if _, err := authenticateMessage.Unmarshal(inner); err != nil {
		return fmt.Errorf("failed to parse the NTLM AUTHENTICATE message: %v", err)
	}

	ctx.Authenticate = authenticateMessage
	ctx.AuthenticateMessageBytes = append([]byte(nil), inner...)

	return nil
}

// Verify checks the recorded AUTHENTICATE against the credential for the identity
// it claimed, and on success derives the ExportedSessionKey.
//
// It reports ErrCaptureOnly when no CredentialLookup is configured,
// ErrUnknownIdentity when the lookup has no credential, ErrBadResponse when the
// NT response does not verify, and ErrBadMIC when the response verifies but the
// message integrity code does not.
//
// Returns:
//   - nil when the authentication is genuine, with SessionKey populated
func (ctx *AcceptContext) Verify() error {
	if ctx.Authenticate == nil {
		return ErrNoAuthenticate
	}
	if ctx.CredentialLookup == nil {
		return ErrCaptureOnly
	}

	domain, username, _ := ctx.Identity()

	ntHash, ok := ctx.CredentialLookup(domain, username)
	if !ok {
		return ErrUnknownIdentity
	}

	// The identity folded into NTOWFv2 must be the one the client sent, byte for
	// byte: NTLMv2 uppercases only the username and takes the domain as-is, so a
	// normalized domain here would recompute a different key and reject every
	// client whose domain was not already upper-case.
	//
	// The client challenge is not part of verification — it reaches the verifier
	// inside the blob, which is hashed as received — so a zero value is passed.
	v2, err := ntlmv2.NewNTLMv2CtxWithNTHash(domain, username, ntHash, ctx.ServerChallenge, [8]byte{})
	if err != nil {
		return fmt.Errorf("failed to derive the response key: %v", err)
	}

	ntResponse := ctx.Authenticate.NtChallengeResponse
	if !ntlmv2.VerifyNTChallengeResponse(v2.ResponseKeyNT[:], ctx.ServerChallenge, ntResponse) {
		return ErrBadResponse
	}

	sessionKey, err := ctx.exportedSessionKey(v2.ResponseKeyNT[:], ntResponse)
	if err != nil {
		return err
	}

	// A client that carries a MIC has committed to the whole exchange, so a
	// mismatch means the NEGOTIATE or CHALLENGE was altered in flight even though
	// the response itself is genuine.
	if ctx.Authenticate.NeedsMIC {
		if !ctx.Authenticate.VerifyMIC(sessionKey, ctx.NegotiateMessageBytes, ctx.ChallengeMessageBytes, ctx.AuthenticateMessageBytes) {
			return ErrBadMIC
		}
	}

	ctx.SessionKey = sessionKey
	ctx.Verified = true

	return nil
}

// exportedSessionKey derives the ExportedSessionKey for a verified response.
//
// For NTLMv2 the KeyExchangeKey is the SessionBaseKey ([MS-NLMP] 3.4.5.1). When
// the client negotiated NTLMSSP_NEGOTIATE_KEY_EXCH it chose its own session key,
// sealed it under the KeyExchangeKey with RC4 and sent it in the AUTHENTICATE, so
// the acceptor has to unseal it ([MS-NLMP] 3.1.5.1.2); otherwise the two keys are
// the same.
func (ctx *AcceptContext) exportedSessionKey(responseKeyNT, ntResponse []byte) ([]byte, error) {
	keyExchangeKey := ntlmv2.SessionBaseKeyFromResponse(responseKeyNT, ntResponse)
	if len(keyExchangeKey) == 0 {
		return nil, fmt.Errorf("failed to derive the session base key from the NT challenge response")
	}

	if !ctx.Authenticate.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_KEY_EXCH) {
		return keyExchangeKey, nil
	}

	encrypted := ctx.Authenticate.EncryptedRandomSessionKey
	if len(encrypted) != 16 {
		return nil, fmt.Errorf("client negotiated key exchange but sent a %d-byte EncryptedRandomSessionKey, want 16", len(encrypted))
	}

	cipher, err := rc4.NewRC4WithKey(keyExchangeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise RC4 for the key exchange: %v", err)
	}
	exported := make([]byte, len(encrypted))
	cipher.XORKeyStream(exported, encrypted)

	return exported, nil
}

// Identity reports the domain, username and workstation the client claimed in its
// AUTHENTICATE, decoded according to the character set the exchange negotiated.
//
// The values are what the client asserted, not anything verified: they are
// meaningful for a capture record before Verify runs, and remain unvalidated
// input until it succeeds.
//
// Returns:
//   - The claimed domain, username and workstation, empty if no AUTHENTICATE
//     has been accepted
func (ctx *AcceptContext) Identity() (domain, username, workstation string) {
	if ctx.Authenticate == nil {
		return "", "", ""
	}

	decode := func(b []byte) string {
		if ctx.Authenticate.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_UNICODE) {
			return utf16.DecodeUTF16LE(b)
		}
		return string(b)
	}

	return decode(ctx.Authenticate.DomainName),
		decode(ctx.Authenticate.UserName),
		decode(ctx.Authenticate.Workstation)
}

// CapturedResponse renders the recorded AUTHENTICATE as a crackable response.
//
// Exactly one of the two is returned, chosen by the length of the NT challenge
// response: 24 bytes is an NTLMv1 response, anything longer is NTLMv2. A response
// that is neither yields two nils, as does a context with no AUTHENTICATE.
//
// Returns:
//   - The NTLMv1 response, or nil
//   - The NTLMv2 response, or nil
func (ctx *AcceptContext) CapturedResponse() (*ntlmv1.NTLMv1Response, *ntlmv2.NTLMv2Response) {
	if ctx.Authenticate == nil {
		return nil, nil
	}

	domain, username, _ := ctx.Identity()
	ntResponse := ctx.Authenticate.NtChallengeResponse

	var lmResponse [24]byte
	copy(lmResponse[:], ctx.Authenticate.LmChallengeResponse)

	switch {
	case len(ntResponse) == 24:
		var nt [24]byte
		copy(nt[:], ntResponse)
		return ntlmv1.NewNTLMv1Response(username, domain, ctx.ServerChallenge, lmResponse, nt), nil

	case len(ntResponse) > 24:
		return nil, ntlmv2.NewNTLMv2Response(username, domain, ctx.ServerChallenge, lmResponse, ntResponse)
	}

	return nil, nil
}

// GetSessionKey returns the ExportedSessionKey derived by a successful Verify, or
// nil if verification has not succeeded.
//
// Returns:
//   - The session key, or nil
func (ctx *AcceptContext) GetSessionKey() []byte {
	return ctx.SessionKey
}

// extractNTLMMessage pulls the NTLMSSP message out of whatever framing a client
// used. Three shapes appear in practice: a bare NTLMSSP message, a continuation
// token (a SecurityBlob wrapping a NegTokenResp), and an initial token (a
// GSS-wrapped NegTokenInit).
func extractNTLMMessage(token []byte) ([]byte, error) {
	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	// A bare NTLMSSP message needs no unwrapping.
	if bytes.HasPrefix(token, header.NTLM_SIGNATURE[:]) {
		return token, nil
	}

	// A continuation token: SecurityBlob wrapping a NegTokenResp.
	blob := &SecurityBlob{}
	if _, err := blob.Unmarshal(token); err == nil {
		resp := NegTokenResp{}
		if _, err := resp.Unmarshal(blob.Data); err == nil && len(resp.ResponseToken) > 0 {
			return resp.ResponseToken, nil
		}
	}

	// An initial token: GSS-wrapped NegTokenInit, or a GSS-wrapped NegTokenResp.
	if inner, err := ExtractNTLMToken(token); err == nil {
		return inner, nil
	}

	return nil, errors.New("no NTLM message found in the token")
}

// expectMessageType checks the NTLMSSP signature and message type before a
// message is parsed, so a client cannot have one message type decoded as another.
func expectMessageType(message []byte, want types.MessageType) error {
	if len(message) < 12 {
		return fmt.Errorf("NTLM message is %d bytes, too short to carry a header", len(message))
	}
	if !bytes.Equal(message[:8], header.NTLM_SIGNATURE[:]) {
		return fmt.Errorf("NTLM message does not carry the NTLMSSP signature")
	}

	messageHeader := &header.Header{}
	if _, err := messageHeader.Unmarshal(message); err != nil {
		return fmt.Errorf("failed to parse the NTLM message header: %v", err)
	}
	if messageHeader.MessageType != want {
		return fmt.Errorf("NTLM message type is %d, want %d", messageHeader.MessageType, want)
	}

	return nil
}

// CompletionToken builds the token that closes a SPNEGO exchange: a NegTokenResp
// reporting accept-completed, wrapped in a SecurityBlob.
//
// This is not optional padding on a successful logon. RFC 4178 section 4.2.2 has
// the acceptor report the outcome of the negotiation in negState, and a mechanism
// that has finished is reported as accept-completed(0); an initiator whose state
// machine is still waiting for that token treats an empty final blob as a
// protocol violation and abandons the session, even though the server considered
// the logon successful. So the final leg carries this rather than nothing.
//
// No responseToken accompanies it: NTLM's last message is the client's
// AUTHENTICATE, so there is nothing left for the server to send. The mechanism is
// named again so an initiator that tracks which mechanism was settled on can
// confirm it.
//
// Returns:
//   - The marshalled SecurityBlob
//   - An error if it cannot be built
func (ctx *AcceptContext) CompletionToken() ([]byte, error) {
	negTokenResp := NewNegTokenResp(NegStateAcceptCompleted, NtlmOID, nil)
	marshalledNegTokenResp, err := negTokenResp.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the completing NegTokenResp: %v", err)
	}

	securityBlob := SecurityBlob{Data: marshalledNegTokenResp}
	marshalledSecurityBlob, err := securityBlob.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the completing SecurityBlob: %v", err)
	}

	return marshalledSecurityBlob, nil
}
