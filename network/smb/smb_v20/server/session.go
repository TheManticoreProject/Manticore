package server

import (
	"bytes"
	"encoding/asn1"
	"errors"
	"net"
	"strings"
	"sync/atomic"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/header"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// Session is an authenticated security context on a connection.
type Session struct {
	id          uint64
	established bool

	Domain      string
	Username    string
	IsGuest     bool
	IsAnonymous bool

	signingKey      []byte
	encryptionKey   []byte // server to client
	decryptionKey   []byte // client to server
	signingRequired bool
	encryptData     bool
	nonce           atomic.Uint64

	// preauthHash is the SMB 3.1.1 hash the session's keys are derived from.
	preauthHash []byte
	// ntlm is the NTLM exchange awaiting its AUTHENTICATE.
	ntlm *spnego.AcceptContext

	trees map[uint32]*Tree
}

// identity is what an authentication established.
type identity struct {
	domain, username string
	sessionKey       []byte
	guest, anonymous bool
}

// sessionSetup answers SESSION_SETUP (MS-SMB2 3.3.5.5).
func (c *Connection) sessionSetup(r *request, resp *response) nt_status.NT_STATUS {
	req := r.command.(*commands.SessionSetupRequest)
	if req.Flags&commands.SMB2_SESSION_FLAG_BINDING != 0 {
		return nt_status.NT_STATUS_REQUEST_NOT_ACCEPTED
	}

	session := c.sessions[r.header.SessionId]
	switch {
	case r.header.SessionId == 0:
		if len(c.sessions) >= c.server.config.MaxSessionsPerConnection {
			return nt_status.NT_STATUS_INSUFFICIENT_RESOURCES
		}
		session = &Session{id: c.server.newSessionID(), preauthHash: c.preauthHash, trees: make(map[uint32]*Tree)}
		c.sessions[session.id] = session
	case session == nil:
		return nt_status.NT_STATUS_USER_SESSION_DELETED
	case session.established:
		// Re-authentication is not served: the client sets up a new session.
		return nt_status.NT_STATUS_REQUEST_NOT_ACCEPTED
	}
	resp.session = session

	if c.dialect == dialects.SMB2_DIALECT_3_1_1 {
		session.preauthHash = security.PreauthUpdate(session.preauthHash, r.raw)
	}

	token, who, status := c.authenticate(session, req.SecurityBuffer)
	setup := commands.NewSessionSetupResponse()
	setup.SecurityBuffer = token
	resp.command = setup

	switch status {
	case nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED:
		if c.dialect == dialects.SMB2_DIALECT_3_1_1 {
			resp.onSent = func(wire []byte) { session.preauthHash = security.PreauthUpdate(session.preauthHash, wire) }
		}
		return status
	case nt_status.NT_STATUS_SUCCESS:
		if status = c.establish(session, req, who, setup); status.IsSuccess() {
			resp.sign = true
			return status
		}
	}
	delete(c.sessions, session.id)
	return status
}

// authenticate runs one leg of the SPNEGO exchange and returns the token to
// answer with. It completes with an identity, or asks for another leg.
func (c *Connection) authenticate(session *Session, token []byte) ([]byte, *identity, nt_status.NT_STATUS) {
	if session.ntlm != nil {
		return c.finishNTLM(session, token)
	}

	if init, err := spnego.ParseNegTokenInit(token); err == nil && len(init.MechToken) > 0 && !isNTLM(init.MechToken) {
		if c.server.config.Kerberos != nil && len(init.MechTypes) > 0 {
			return c.acceptKerberos(init)
		}
		// The optimistic token is for a mechanism not served: steer the client to NTLM.
		redirect, err := completionToken(spnego.NegStateAcceptIncomplete, spnego.NtlmOID, nil)
		if err != nil {
			return nil, nil, nt_status.NT_STATUS_INTERNAL_ERROR
		}
		return redirect, nil, nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED
	}

	config := c.server.config
	info, err := targetinfo.BuildServerTargetInfo(config.ServerName, config.DomainName, config.DNSComputerName, config.DNSDomainName, nil)
	if err != nil {
		return nil, nil, nt_status.NT_STATUS_INTERNAL_ERROR
	}
	accept := spnego.NewAcceptContext(config.DomainName, challenge.TargetTypeDomain, info)
	accept.CredentialLookup = config.Authenticator
	challengeToken, err := accept.AcceptNegotiateToken(token)
	if err != nil {
		logger.Debugf("SMB2 server: unusable NTLM NEGOTIATE from %s: %v", c.remote, err)
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}
	session.ntlm = accept
	return challengeToken, nil, nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED
}

// finishNTLM verifies the AUTHENTICATE that completes an NTLM exchange.
func (c *Connection) finishNTLM(session *Session, token []byte) ([]byte, *identity, nt_status.NT_STATUS) {
	accept := session.ntlm
	session.ntlm = nil
	if err := accept.AcceptAuthenticateToken(token); err != nil {
		return nil, nil, nt_status.NT_STATUS_INVALID_PARAMETER
	}

	who := &identity{}
	who.domain, who.username, _ = accept.Identity()
	switch err := accept.Verify(); {
	case err == nil:
		who.sessionKey = accept.GetSessionKey()
	case errors.Is(err, spnego.ErrUnknownIdentity), errors.Is(err, spnego.ErrCaptureOnly):
		// An identity the server cannot verify is admitted only as a guest.
		who.anonymous = who.username == "" && len(accept.Authenticate.NtChallengeResponse) == 0
		who.guest = !who.anonymous
	default:
		logger.Debugf("SMB2 server: refusing %s\\%s from %s: %v", who.domain, who.username, c.remote, err)
		return nil, nil, nt_status.NT_STATUS_LOGON_FAILURE
	}

	done, err := accept.CompletionToken()
	if err != nil {
		return nil, nil, nt_status.NT_STATUS_INTERNAL_ERROR
	}
	return done, who, nt_status.NT_STATUS_SUCCESS
}

// acceptKerberos verifies a Kerberos AP-REQ, which completes in one leg.
func (c *Connection) acceptKerberos(init *spnego.NegTokenInit) ([]byte, *identity, nt_status.NT_STATUS) {
	options := *c.server.config.Kerberos
	if addr, ok := c.remote.(*net.TCPAddr); ok {
		options.ClientAddress = addr.IP
	}
	apRep, ctx, err := gssapi.AcceptSecContext(init.MechToken, options)
	if err != nil {
		logger.Debugf("SMB2 server: refusing a Kerberos logon from %s: %v", c.remote, err)
		return nil, nil, nt_status.NT_STATUS_LOGON_FAILURE
	}

	who := &identity{sessionKey: ctx.SessionKey}
	if len(ctx.SubKey) > 0 {
		who.sessionKey = ctx.SubKey
	}
	name, realm := ctx.ClientPrincipal()
	who.domain, who.username = realm, strings.Join(name.NameString, "/")

	done, err := completionToken(spnego.NegStateAcceptCompleted, init.MechTypes[0], apRep)
	if err != nil {
		return nil, nil, nt_status.NT_STATUS_INTERNAL_ERROR
	}
	return done, who, nt_status.NT_STATUS_SUCCESS
}

// establish turns an authenticated identity into a usable session, deriving its
// keys and applying the signing and encryption policy.
func (c *Connection) establish(session *Session, req *commands.SessionSetupRequest, who *identity, setup *commands.SessionSetupResponse) nt_status.NT_STATUS {
	config := c.server.config
	keyless := who.guest || who.anonymous
	switch {
	case who.anonymous && !config.AllowAnonymous, who.guest && !config.AllowGuest:
		return nt_status.NT_STATUS_LOGON_FAILURE
	case keyless && (config.RequireSigning || config.RequireEncryption):
		// A session without a key could not carry a single request.
		return nt_status.NT_STATUS_LOGON_FAILURE
	case config.RequireEncryption && (c.cipher == 0 || !security.IsSMB3(c.dialect)):
		return nt_status.NT_STATUS_ACCESS_DENIED
	}

	session.established = true
	session.Domain, session.Username = who.domain, who.username
	session.IsGuest, session.IsAnonymous = who.guest, who.anonymous
	switch {
	case who.anonymous:
		setup.SessionFlags = commands.SMB2_SESSION_FLAG_IS_NULL
	case who.guest:
		setup.SessionFlags = commands.SMB2_SESSION_FLAG_IS_GUEST
	default:
		// The SMB2 session key is the first 16 bytes of the GSS key, zero-padded.
		// The full GSS key (Session.FullSessionKey) derives the SMB 3.1.1 AES-256
		// cipher keys (MS-SMB2 3.3.5.5.3).
		key := make([]byte, 16)
		copy(key, who.sessionKey)
		session.signingKey = key
		if security.IsSMB3(c.dialect) {
			keys := security.DeriveKeysWithFullSessionKey(key, who.sessionKey, c.dialect, session.preauthHash, c.cipher)
			session.signingKey = keys.Signing
			if c.cipher != 0 {
				session.encryptionKey, session.decryptionKey = keys.ServerToClient, keys.ClientToServer
			}
		}
		session.signingRequired = config.RequireSigning || req.SecurityMode.IsSigningRequired()
		if config.RequireEncryption {
			session.encryptData = true
			setup.SessionFlags = commands.SMB2_SESSION_FLAG_ENCRYPT_DATA
		}
	}

	admission := ""
	switch {
	case who.anonymous:
		admission = " (anonymous)"
	case who.guest:
		admission = " (guest)"
	}
	logger.Infof("SMB2 server: %s authenticated as %s\\%s%s", c.remote, session.Domain, session.Username, admission)
	return nt_status.NT_STATUS_SUCCESS
}

// logoff answers LOGOFF, releasing everything the session holds.
func (c *Connection) logoff(r *request, resp *response) nt_status.NT_STATUS {
	for _, tree := range r.session.trees {
		c.disconnectTree(r.session, tree)
	}
	delete(c.sessions, r.session.id)
	resp.command = commands.NewLogoffResponse()
	return nt_status.NT_STATUS_SUCCESS
}

// isNTLM reports whether a mechanism token is a bare NTLMSSP message.
func isNTLM(token []byte) bool {
	return bytes.HasPrefix(token, header.NTLM_SIGNATURE[:])
}

// completionToken builds a NegTokenResp wrapped as a SESSION_SETUP security blob.
func completionToken(state spnego.NegState, mech asn1.ObjectIdentifier, token []byte) ([]byte, error) {
	resp, err := spnego.NewNegTokenResp(state, mech, token).Marshal()
	if err != nil {
		return nil, err
	}
	blob := spnego.SecurityBlob{Data: resp}
	return blob.Marshal()
}
