package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// SessionSetupNegotiateBlob sends the first SESSION_SETUP leg carrying a
// caller-supplied NTLM NEGOTIATE token and returns the server's raw CHALLENGE
// token together with the SessionId the server assigned.
//
// Unlike SessionSetup, it computes nothing: the token is passed through
// unchanged and no credential is held. It is the outbound opening leg of an NTLM
// relay, where the NEGOTIATE comes from a third party whose exchange is being
// carried. Pair it with SessionSetupAuthenticateBlob, feeding the returned
// SessionId back.
//
// Returns:
//   - The server's CHALLENGE (type-2) token, to hand back to the third party
//   - The SessionId the server assigned, for the completing leg
//   - An error if the transport is down or the server did not ask for a second leg
func (c *Client) SessionSetupNegotiateBlob(negotiateToken []byte) (challengeToken []byte, sessionId uint64, err error) {
	if !c.Transport.IsConnected() {
		return nil, 0, fmt.Errorf("transport is not connected")
	}

	req := commands.NewSessionSetupRequest()
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.SecurityBuffer = negotiateToken

	msg := c.newRequest(req)
	msg.Header.SessionId = 0

	resp, err := c.sendReceive(msg, "SessionSetupRelay(negotiate)")
	if err != nil {
		return nil, 0, err
	}
	if status := statusFromResponse(resp); status != ntStatusMoreProcessingRequired {
		return nil, 0, fmt.Errorf("unexpected SESSION_SETUP status: %s", formatNTStatus(status))
	}

	challengeResp, ok := resp.Command.(*commands.SessionSetupResponse)
	if !ok {
		return nil, 0, fmt.Errorf("unexpected SESSION_SETUP response command: %T", resp.Command)
	}

	return challengeResp.SecurityBuffer, resp.Header.SessionId, nil
}

// SessionSetupAuthenticateBlob sends the second SESSION_SETUP leg carrying a
// caller-supplied NTLM AUTHENTICATE token for the SessionId returned by
// SessionSetupNegotiateBlob, completing the outbound leg of an NTLM relay.
//
// Because this client holds no credential, it derives no key: on success the
// session is installed unsigned, exactly as a guest session is (MS-SMB2
// 3.2.5.3.1). The relayed session is therefore usable only against a server that
// does not require signing — the inherent limit of a relay, which cannot sign for
// an identity whose key it never learns.
//
// Returns:
//   - An error if the transport is down or the server rejected the AUTHENTICATE
func (c *Client) SessionSetupAuthenticateBlob(sessionId uint64, authenticateToken []byte) error {
	if !c.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}

	// No key is derived, so the session carries a zero key and never signs. The
	// AUTHENTICATE leg is unsigned regardless, so installing the session before
	// sending it matches the credential path without arming signatures.
	session := &Session{
		Client:        c,
		SessionId:     sessionId,
		SessionKey:    make([]byte, 16),
		SigningActive: false,
	}
	session.SigningKey = session.SessionKey
	c.Session = session

	req := commands.NewSessionSetupRequest()
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.SecurityBuffer = authenticateToken

	msg := c.newRequest(req)
	msg.Header.SessionId = sessionId

	resp, err := c.sendReceive(msg, "SessionSetupRelay(authenticate)")
	if err != nil {
		c.Session = nil
		return err
	}
	if status := statusFromResponse(resp); status != 0x00000000 {
		c.Session = nil
		return fmt.Errorf("session setup failed: %s", formatNTStatus(status))
	}

	if resp2Cmd, ok := resp.Command.(*commands.SessionSetupResponse); ok {
		session.IsGuest = resp2Cmd.SessionFlags&commands.SMB2_SESSION_FLAG_IS_GUEST != 0
		session.IsNull = resp2Cmd.SessionFlags&commands.SMB2_SESSION_FLAG_IS_NULL != 0
	}

	if c.Connection.SessionTable == nil {
		c.Connection.SessionTable = make(map[uint64]*Session)
	}
	c.Connection.SessionTable[sessionId] = session

	return nil
}
