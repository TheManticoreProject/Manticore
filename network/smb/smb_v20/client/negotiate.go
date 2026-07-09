package client

import (
	"crypto/rand"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// preauthSaltLength is the length of the salt the client places in the SMB 3.1.1
// SMB2_PREAUTH_INTEGRITY_CAPABILITIES negotiate context; Windows uses 32 bytes.
const preauthSaltLength = 32

// Negotiate performs the SMB2 NEGOTIATE exchange, offering the full range of
// SMB 2.0.2 through 3.1.1 dialects and capturing the server's chosen dialect,
// security mode, capabilities, maximum transfer sizes, GUID, timestamps, and
// GSS security buffer.
//
// For the SMB 3.1.1 dialect the request carries the pre-authentication
// integrity (SHA-512) and encryption (AES-128-GCM / AES-128-CCM) negotiate
// contexts, and the running pre-auth integrity hash is seeded from the exact
// NEGOTIATE request and response bytes (MS-SMB2 3.1.4.2).
//
// NEGOTIATE uses MessageId 0 and SessionId 0, as required by the spec.
func (c *Client) Negotiate() error {
	req := commands.NewNegotiateRequest()
	req.AddDialect(dialects.SMB2_DIALECT_2_0_2)
	req.AddDialect(dialects.SMB2_DIALECT_2_1_0)
	req.AddDialect(dialects.SMB2_DIALECT_3_0_0)
	req.AddDialect(dialects.SMB2_DIALECT_3_0_2)
	req.AddDialect(dialects.SMB2_DIALECT_3_1_1)
	req.ClientGuid = c.ClientGuid
	// Advertise signing support and the SMB 3.x capabilities (large MTU,
	// encryption) so the server may negotiate up to 3.1.1.
	req.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	req.Capabilities = capabilities.SMB2_GLOBAL_CAP_LARGE_MTU | capabilities.SMB2_GLOBAL_CAP_ENCRYPTION

	// SMB 3.1.1 negotiate contexts: pre-auth integrity (SHA-512 + random salt)
	// and encryption ciphers in preference order (AES-128-GCM, then AES-128-CCM).
	salt := make([]byte, preauthSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("negotiate: failed to generate pre-auth salt: %w", err)
	}
	req.Contexts = []*commands.NegotiateContext{
		commands.NewPreauthIntegrityContext(salt),
		commands.NewEncryptionContext([]uint16{
			commands.SMB2_ENCRYPTION_AES128_GCM,
			commands.SMB2_ENCRYPTION_AES128_CCM,
		}),
	}

	// Seed the pre-auth integrity hash with 64 zero bytes; it is folded with the
	// NEGOTIATE request and response bytes below.
	c.Connection.PreauthIntegrityHashValue = make([]byte, preauthHashLength)

	msg := c.newRequest(req)
	// NEGOTIATE is sent before any session exists; SessionId MUST be 0.
	msg.Header.SessionId = 0

	response, err := c.sendReceive(msg, "Negotiate")
	if err != nil {
		return err
	}

	// Fold the exact request and response wire bytes into the pre-auth hash.
	c.Connection.PreauthIntegrityHashValue = preauthUpdate(c.Connection.PreauthIntegrityHashValue, c.lastSentBytes)
	c.Connection.PreauthIntegrityHashValue = preauthUpdate(c.Connection.PreauthIntegrityHashValue, c.lastRecvBytes)

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

	// SMB 3.1.1: record the server's chosen cipher and pre-auth hash algorithm
	// from the returned negotiate contexts.
	if negotiateResponse.DialectRevision == dialects.SMB2_DIALECT_3_1_1 {
		c.Connection.Cipher = commands.SelectedCipher(negotiateResponse.Contexts)
		c.Connection.PreauthIntegrityHashId = commands.SelectedPreauthHash(negotiateResponse.Contexts)
	}

	if c.Connection.MessageId == 0 {
		c.Connection.MessageId = 1
	}
}
