package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// Negotiate performs the SMB2 NEGOTIATE exchange, requesting the SMB 2.0.2
// dialect and capturing the server's security mode, capabilities, maximum
// transfer sizes, GUID, timestamps, and GSS security buffer.
//
// NEGOTIATE uses MessageId 0 and SessionId 0, as required by the spec.
func (c *Client) Negotiate() error {
	req := commands.NewNegotiateRequest()
	req.AddDialect(dialects.SMB2_DIALECT_2_0_2)
	req.ClientGuid = c.ClientGuid
	// Advertise that the client supports signing.
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED

	msg := c.newRequest(req)
	// NEGOTIATE is sent before any session exists; SessionId MUST be 0.
	msg.Header.SessionId = 0

	response, err := c.sendReceive(msg, "Negotiate")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("negotiate failed: 0x%08x", status)
	}

	negotiateResponse, ok := response.Command.(*commands.NegotiateResponse)
	if !ok {
		return fmt.Errorf("unexpected negotiate response command: %T", response.Command)
	}

	c.ApplyNegotiateResponse(negotiateResponse)

	return nil
}

// ApplyNegotiateResponse records the negotiated connection state — selected
// dialect, server security mode, capabilities, maximum transfer sizes, GUID,
// timestamps, and GSS security buffer — from an SMB2 NEGOTIATE response.
//
// It is shared by Negotiate, which exchanges the response on the transport, and
// by the generic SMB client (network/smb/client), which may hand over a response
// already exchanged as part of a multi-protocol negotiate. Because NEGOTIATE
// occupies MessageId 0, the next request on the connection uses MessageId 1; this
// is set here for the handoff path (in the Negotiate path the MessageId has
// already advanced past 0).
func (c *Client) ApplyNegotiateResponse(negotiateResponse *commands.NegotiateResponse) {
	server := c.Connection.Server
	c.Connection.Dialect = negotiateResponse.DialectRevision
	server.SecurityMode = negotiateResponse.SecurityMode
	server.Capabilities = negotiateResponse.Capabilities
	server.MaxTransactSize = negotiateResponse.MaxTransactSize
	server.MaxReadSize = negotiateResponse.MaxReadSize
	server.MaxWriteSize = negotiateResponse.MaxWriteSize
	server.ServerGuid = negotiateResponse.ServerGuid
	server.SystemTime = negotiateResponse.SystemTime
	server.ServerStartTime = negotiateResponse.ServerStartTime
	server.SecurityBuffer = negotiateResponse.SecurityBuffer

	if c.Connection.MessageId == 0 {
		c.Connection.MessageId = 1
	}
}
