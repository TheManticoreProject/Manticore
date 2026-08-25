package server

import (
	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// provisionalUID is the user identifier returned alongside the challenge, before
// anything has been authenticated. The client echoes it on the second leg so the
// server can match the two halves of one exchange.
//
// It is a fixed value because this phase serves one authentication attempt per
// connection: a session table, and with it real identifier allocation, arrives
// with the phase that establishes authenticated sessions.
const provisionalUID = 0x0800

// handleSessionSetupAndx answers SMB_COM_SESSION_SETUP_ANDX.
//
// Extended security makes this a two-leg exchange. The first request carries the
// client's NTLM NEGOTIATE, and is answered with a CHALLENGE and
// STATUS_MORE_PROCESSING_REQUIRED. The second carries the AUTHENTICATE, which is
// recorded and then refused, because verifying it needs a credential store that
// arrives with a later phase.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/b7b0e2e5-4b62-4b0f-b2ba-6dfe0c4e6d5f
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
	// it is expecting to send a password or a bare NTLM response against the
	// challenge from the negotiate response, which this server does not issue.
	if !conn.ExtendedSecurity || len(request.SecurityBlob) == 0 {
		logger.Debugf("SMB1 server: %s attempted a session setup without extended security", conn.Remote)
		return nt_status.NT_STATUS_NOT_IMPLEMENTED
	}

	// Record what the client says it can handle. It bounds what the server may
	// send back, so it is worth keeping even before the session exists.
	conn.ClientMaxBufferSize = uint32(request.MaxBufferSize)
	conn.ClientCapabilities = request.Capabilities

	if conn.Accept == nil {
		return conn.beginAuthentication(w, request)
	}
	return conn.finishAuthentication(w, request)
}

// beginAuthentication answers the first leg: it consumes the client's NEGOTIATE
// and returns the CHALLENGE.
func (c *Connection) beginAuthentication(w ResponseWriter, request *commands.SessionSetupAndxRequest) nt_status.NT_STATUS {
	targetInfo, err := c.Server.targetInfo()
	if err != nil {
		logger.Errorf("SMB1 server: failed to build the TargetInfo for %s: %v", c.Remote, err)
		return nt_status.NT_STATUS_UNSUCCESSFUL
	}

	// The NetBIOS domain names the target, since that is what a client resolves
	// an account against.
	accept := spnego.NewAcceptContext(c.Server.config.DomainName, challenge.TargetTypeDomain, targetInfo)

	challengeBlob, err := accept.AcceptNegotiateToken([]byte(request.SecurityBlob))
	if err != nil {
		logger.Debugf("SMB1 server: could not answer the NTLM NEGOTIATE from %s: %v", c.Remote, err)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	c.Accept = accept

	logger.Debugf("SMB1 server: issued an NTLM challenge %x to %s", accept.ServerChallenge, c.Remote)

	response := c.newSessionSetupResponse()
	response.SecurityBlob = []types.UCHAR(challengeBlob)
	response.SecurityBlobLength = types.USHORT(len(challengeBlob))

	// The challenge leg reports that the exchange is unfinished, and carries the
	// identifier the client echoes on the second leg.
	w.SetResponseUID(provisionalUID)
	if err := w.WriteResponseWithStatus(response, nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED); err != nil {
		logger.Debugf("SMB1 server: failed to send the NTLM challenge to %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// finishAuthentication answers the second leg: it records the AUTHENTICATE and
// then decides.
//
// The decision is currently always a refusal, because verifying a response needs
// a credential store this phase does not have. Recording first means the material
// is not lost by the refusal, which is what a capture handler relies on.
func (c *Connection) finishAuthentication(w ResponseWriter, request *commands.SessionSetupAndxRequest) nt_status.NT_STATUS {
	if err := c.Accept.AcceptAuthenticateToken([]byte(request.SecurityBlob)); err != nil {
		logger.Debugf("SMB1 server: could not read the NTLM AUTHENTICATE from %s: %v", c.Remote, err)
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}

	domain, username, workstation := c.Accept.Identity()
	logger.Debugf("SMB1 server: %s authenticated as %s\\%s from %q", c.Remote, domain, username, workstation)

	// Verify reports why it could not confirm the authentication. With no
	// credential store the answer is always ErrCaptureOnly, and the honest reply
	// is a logon failure: it is what a client retries against, and it does not
	// pretend the server could have served the session.
	if err := c.Accept.Verify(); err != nil {
		logger.Debugf("SMB1 server: refusing %s\\%s from %s: %v", domain, username, c.Remote, err)
	}

	return nt_status.NT_STATUS_LOGON_FAILURE
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
// over the whole exchange, and while the acceptor can verify a MIC, requiring one
// gains nothing for a server that is not yet completing sessions and costs
// interoperability with clients that handle it poorly.
func (s *Server) targetInfo() ([]byte, error) {
	return targetinfo.BuildServerTargetInfo(
		s.config.ServerName,
		s.config.DomainName,
		s.config.DNSComputerName,
		s.config.DNSDomainName,
		nil,
	)
}
