package server

import (
	"errors"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// authenticateRequestSequenceNumber is the sequence number the AUTHENTICATE
// request is signed at, and so the number signing is bootstrapped from. Its
// response takes the number above, and the first request after it the one above
// that.
const authenticateRequestSequenceNumber = 0

// handleSessionSetupAndx answers SMB_COM_SESSION_SETUP_ANDX.
//
// Extended security makes this a two-leg exchange, and the legs are told apart by
// the UID: a client opens one with UID 0 and echoes the identifier the server
// assigned on the second leg. That is what allows a connection to carry several
// sessions — once one identity is established, the next setup starts a fresh
// exchange rather than being mistaken for the tail of the previous one.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/8c9d1b1e-3b0e-4c1d-9dcb-e5e0f6b5e5a5
func handleSessionSetupAndx(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.SessionSetupAndxRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	// A dialect has to be agreed before a session can be set up on it.
	if !conn.Negotiated {
		logger.Debugf("SMB1 server: %s sent SESSION_SETUP_ANDX before negotiating", conn.Remote)
		return nt_status.NT_STATUS_INVALID_SMB
	}

	// Only the extended-security path is served. A client that did not negotiate
	// it expects to send a password or a bare NTLM response against a challenge
	// from the negotiate response, which this server does not issue.
	if !conn.ExtendedSecurity || len(request.SecurityBlob) == 0 {
		logger.Debugf("SMB1 server: %s attempted a session setup without extended security", conn.Remote)
		return nt_status.NT_STATUS_NOT_IMPLEMENTED
	}

	// Record what the client says it can handle: it bounds what may be sent back.
	conn.ClientMaxBufferSize = uint32(request.MaxBufferSize)
	conn.ClientCapabilities = request.Capabilities

	uid := uint16(req.Header.UID)
	switch {
	case uid == 0:
		return conn.beginAuthentication(w, request)
	case conn.PendingAuth(uid) != nil:
		return conn.finishAuthentication(w, req, request, uid)
	default:
		// A UID naming neither a pending exchange nor zero is either an
		// established session being reused for a setup, or a value the client
		// invented.
		logger.Debugf("SMB1 server: %s continued a session setup on unknown UID 0x%04X", conn.Remote, uid)
		return nt_status.NT_STATUS_SMB_BAD_UID
	}
}

// beginAuthentication answers the first leg: it consumes the client's NEGOTIATE,
// assigns a UID and returns the CHALLENGE.
func (c *Connection) beginAuthentication(w ResponseWriter, request *commands.SessionSetupAndxRequest) nt_status.NT_STATUS {
	uid, err := c.uids.Allocate()
	if err != nil {
		logger.Warnf("SMB1 server: refusing a session setup from %s: %v", c.Remote, err)
		return nt_status.NT_STATUS_TOO_MANY_SESSIONS
	}

	targetInfo, err := c.Server.targetInfo()
	if err != nil {
		logger.Errorf("SMB1 server: failed to build the TargetInfo for %s: %v", c.Remote, err)
		c.uids.Release(uid)
		return nt_status.NT_STATUS_UNSUCCESSFUL
	}

	// The NetBIOS domain names the target, since that is what a client resolves
	// an account against.
	accept := spnego.NewAcceptContext(c.Server.config.DomainName, challenge.TargetTypeDomain, targetInfo)
	accept.CredentialLookup = c.Server.config.Authenticator

	challengeBlob, err := accept.AcceptNegotiateToken([]byte(request.SecurityBlob))
	if err != nil {
		logger.Debugf("SMB1 server: could not answer the NTLM NEGOTIATE from %s: %v", c.Remote, err)
		c.uids.Release(uid)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	c.pendingAuth[uid] = accept

	logger.Debugf("SMB1 server: issued an NTLM challenge %x to %s on UID 0x%04X", accept.ServerChallenge, c.Remote, uid)

	response := c.newSessionSetupResponse()
	response.SecurityBlob = []types.UCHAR(challengeBlob)
	response.SecurityBlobLength = types.USHORT(len(challengeBlob))

	// The challenge leg reports that the exchange is unfinished, and carries the
	// identifier the client echoes on the second leg.
	w.SetResponseUID(uid)
	if err := w.WriteResponseWithStatus(response, nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED); err != nil {
		logger.Debugf("SMB1 server: failed to send the NTLM challenge to %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// finishAuthentication answers the second leg: it records the AUTHENTICATE,
// decides whether to honour it, and on success establishes the session and arms
// signing.
func (c *Connection) finishAuthentication(
	w ResponseWriter,
	req *message.Message,
	request *commands.SessionSetupAndxRequest,
	uid uint16,
) nt_status.NT_STATUS {
	accept := c.pendingAuth[uid]

	// The exchange is over either way. A client that fails starts a new one
	// rather than retrying against the same challenge, which would let it grind
	// candidate passwords against a single nonce.
	defer delete(c.pendingAuth, uid)

	if err := accept.AcceptAuthenticateToken([]byte(request.SecurityBlob)); err != nil {
		logger.Debugf("SMB1 server: could not read the NTLM AUTHENTICATE from %s: %v", c.Remote, err)
		c.uids.Release(uid)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	domain, username, workstation := accept.Identity()
	session := &Session{
		UID:         uid,
		Domain:      domain,
		Username:    username,
		Workstation: workstation,
		Created:     time.Now().UTC(),
	}

	verifyErr := accept.Verify()
	switch {
	case verifyErr == nil:
		session.SessionKey = accept.GetSessionKey()

	case errors.Is(verifyErr, spnego.ErrUnknownIdentity), errors.Is(verifyErr, spnego.ErrCaptureOnly):
		// Either there is no credential store at all, or it does not know this
		// identity. Both are admissible only as a guest, and only where a caller
		// has asked for that.
		anonymous := isAnonymousAttempt(accept)
		if !c.Server.config.AllowGuest || (anonymous && !c.Server.config.AllowAnonymous) {
			logger.Debugf("SMB1 server: refusing %s from %s: %v", session.Account(), c.Remote, verifyErr)
			c.uids.Release(uid)
			return nt_status.NT_STATUS_LOGON_FAILURE
		}
		session.IsGuest = true
		session.IsAnonymous = anonymous

	default:
		// A response that did not verify, or a MIC that did not, is refused
		// regardless of policy: guest access is for an identity the server does
		// not know, not for one whose credential was answered wrongly.
		logger.Debugf("SMB1 server: refusing %s from %s: %v", session.Account(), c.Remote, verifyErr)
		c.uids.Release(uid)
		return nt_status.NT_STATUS_LOGON_FAILURE
	}

	// Decide about signing before committing the session, because a session that
	// cannot meet the policy must not be established at all.
	armSigning, status := c.signingDecision(req, session)
	if status != nt_status.NT_STATUS_SUCCESS {
		c.uids.Release(uid)
		return status
	}

	// Signing belongs to the connection rather than to a session: one key and one
	// sequence carry every session on it. So it is armed once, by the first
	// exchange that can, and a later session setup leaves it alone.
	//
	// Re-keying here would strand every session already established. The
	// connection's sequence would restart at two while the client is far past it,
	// and the key would change under sessions that never asked for a new one, so
	// the next request any of them sent would fail verification and the connection
	// would be dropped.
	if armSigning && !c.SigningActive {
		// Signing is activated by this exchange rather than checked on it. The
		// AUTHENTICATE request's own signature is not verified, and [MS-SMB]
		// section 3.3.5.3 does not ask for it to be: it says only that once the key
		// is acquired "the server MUST sign the SMB_COM_SESSION_SETUP_ANDX
		// response ... by passing in a sequence number of one". The first request
		// the server verifies is therefore the next one, at two.
		//
		// Verifying it would also add nothing. The key is derived from the very
		// message the signature would cover, so anyone able to produce an
		// AUTHENTICATE that verifies can produce its signature as well. What it
		// does do is break real clients: a client that has not yet armed signing
		// puts a placeholder in the field, and refusing that makes signing
		// unusable rather than mandatory.
		c.SigningActive = true
		c.SigningKey = session.SessionKey
		c.ExpectedRequestSequenceNumber = signing.NextRequestSequenceNumber(authenticateRequestSequenceNumber)
		w.SignResponse(session.SessionKey, signing.ResponseSequenceNumber(authenticateRequestSequenceNumber))

		logger.Debugf("SMB1 server: signing armed for %s on UID 0x%04X", c.Remote, uid)
	}
	// On a connection that is already signing, this response needs no signature
	// applied here: handleFrame verified the request against the connection's key
	// and armed the writer with the response sequence that follows it.

	response := c.newSessionSetupResponse()
	if session.IsGuest {
		response.Action = types.USHORT(SMB_SETUP_GUEST)
	}
	// The final leg carries the token that closes the SPNEGO exchange. There is no
	// NTLM message left to send — the client's AUTHENTICATE was the last one — but
	// the negotiation still has to be reported as settled: an initiator waiting for
	// accept-completed treats an empty final blob as a protocol violation and drops
	// the session, however successful the logon was.
	//
	// A client that authenticated without extended security has no SPNEGO exchange
	// to close, so it gets nothing.
	response.SecurityBlob = []types.UCHAR{}
	response.SecurityBlobLength = types.USHORT(0)
	if accept != nil {
		completion, err := accept.CompletionToken()
		if err != nil {
			logger.Warnf("SMB1 server: could not build the completing SPNEGO token for %s: %v", c.Remote, err)
			c.uids.Release(uid)
			return nt_status.NT_STATUS_INTERNAL_ERROR
		}
		response.SecurityBlob = []types.UCHAR(completion)
		response.SecurityBlobLength = types.USHORT(len(completion))
	}

	c.addSession(session)
	logger.Infof("SMB1 server: %s authenticated as %s%s", c.Remote, session.Account(), admissionSuffix(session))

	w.SetResponseUID(uid)
	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to complete the session setup for %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// signingDecision reports whether signing should be armed for a session, and the
// status to refuse it with when the policy cannot be met.
//
// A guest or anonymous session derives no key, so it cannot sign. Under a policy
// that requires signatures such a session is refused rather than established: it
// could not carry a single subsequent request, and granting it would look like
// success while breaking everything after.
func (c *Connection) signingDecision(req *message.Message, session *Session) (bool, nt_status.NT_STATUS) {
	switch c.Server.config.SigningPolicy {
	case SigningRequired:
		if !session.CanSign() {
			logger.Debugf("SMB1 server: refusing %s from %s: signing is required and the session has no key",
				session.Account(), c.Remote)
			return false, nt_status.NT_STATUS_LOGON_FAILURE
		}
		return true, nt_status.NT_STATUS_SUCCESS

	case SigningEnabled:
		// Offered, so it is the client's choice, which it signals with
		// SMB_FLAGS2_SECURITY_SIGNATURE on the request.
		wanted := req.Header.Flags2&flags2.FLAGS2_SECURITY_SIGNATURE != 0
		return wanted && session.CanSign(), nt_status.NT_STATUS_SUCCESS
	}

	return false, nt_status.NT_STATUS_SUCCESS
}

// isAnonymousAttempt reports whether an authentication claimed no identity and
// carried no response, which is what a null session looks like.
func isAnonymousAttempt(accept *spnego.AcceptContext) bool {
	if accept.Authenticate == nil {
		return false
	}
	_, username, _ := accept.Identity()
	return username == "" && len(accept.Authenticate.NtChallengeResponse) == 0
}

// admissionSuffix renders how a session was admitted, for the log line that
// records it. A logon granted as a guest is worth saying so about.
func admissionSuffix(session *Session) string {
	switch {
	case session.IsAnonymous:
		return " (anonymous)"
	case session.IsGuest:
		return " (guest)"
	}
	return ""
}

// handleLogoffAndx answers SMB_COM_LOGOFF_ANDX: it drops the session the request
// arrived on and releases its identifier for reuse.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/2b9d1e6a-1f4e-4a4f-9b60-5a5c5f5d2c9f
func handleLogoffAndx(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	uid := uint16(req.Header.UID)

	session := conn.removeSession(uid)
	if session == nil {
		return nt_status.NT_STATUS_SMB_BAD_UID
	}

	// A logoff releases the trees and handles the session held, rather than
	// leaving them reachable through a UID that no longer names anything.
	conn.closeSessionResources(uid)

	logger.Debugf("SMB1 server: %s logged off %s on UID 0x%04X", conn.Remote, session.Account(), uid)

	// Signing stays armed: it is a property of the connection rather than of one
	// session, and another session on the same connection still expects it.
	if err := w.WriteResponse(commands.NewLogoffAndxResponse()); err != nil {
		logger.Debugf("SMB1 server: failed to answer LOGOFF_ANDX for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// newSessionSetupResponse builds the parts of a session-setup response that do
// not depend on which leg is being answered.
//
// The response's NativeOS and NativeLanMan fields hold wire bytes rather than Go
// strings, so they are encoded here in whichever character set the connection
// negotiated, each with its terminator — the same convention the request side of
// this command uses.
//
// No Pad byte is emitted. The field exists to align NativeOS on a 16-bit
// boundary, but the decoder in this repository's client reads the strings
// immediately after the security blob without skipping one, so a pad byte would
// be decoded as a leading NUL. Whether a real Windows client wants the padding is
// a question for live validation.
func (c *Connection) newSessionSetupResponse() *commands.SessionSetupAndxResponse {
	response := commands.NewSessionSetupAndxResponse()
	response.Pad = []types.UCHAR{}
	response.NativeOS = encodeNativeString(c.Server.config.NativeOS, c.UseUnicode)
	response.NativeLanMan = encodeNativeString(c.Server.config.NativeLanMan, c.UseUnicode)
	return response
}

// encodeNativeString renders an informational string for the wire: UTF-16LE with
// a two-byte terminator when the connection negotiated Unicode, and OEM bytes
// with a one-byte terminator otherwise.
//
// Both strings must be non-empty; strict clients reject a session setup that
// leaves them blank, which is why NewServer defaults them.
func encodeNativeString(value string, useUnicode bool) []types.UCHAR {
	if useUnicode {
		return append(utf16.EncodeUTF16LE(value), 0x00, 0x00)
	}
	return append([]types.UCHAR(value), 0x00)
}

// targetInfo composes the AV_PAIR list advertised in the CHALLENGE.
//
// No MsvAvTimestamp is included. Supplying one obliges the client to carry a MIC
// over the whole exchange; the acceptor can verify one, but requiring it costs
// interoperability with clients that handle it poorly and gains nothing that
// signing does not already give.
func (s *Server) targetInfo() ([]byte, error) {
	return targetinfo.BuildServerTargetInfo(
		s.config.ServerName,
		s.config.DomainName,
		s.config.DNSComputerName,
		s.config.DNSDomainName,
		nil,
	)
}
