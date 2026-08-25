package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
)

// Server signing policy states, recorded in Server.SigningState after negotiation.
const (
	SigningStateDisabled = "disabled"
	SigningStateEnabled  = "enabled"
	SigningStateRequired = "required"
)

// signOutgoing signs a marshalled request in place when signing is active for the
// connection, consuming the next send sequence number. It returns the sequence
// number expected on the matching response and whether the message was signed.
// When signing is inactive it leaves the message untouched and returns (0, false).
func (c *Client) signOutgoing(marshalled []byte) (uint32, bool) {
	if c.Connection == nil || !c.Connection.IsSigningActive {
		return 0, false
	}

	sequenceNumber := c.Connection.ClientNextSendSequenceNumber
	signing.Sign(c.Connection.SigningSessionKey, marshalled, sequenceNumber)

	// The request and its response consume one sequence number each, so the
	// following request skips past both.
	c.Connection.ClientNextSendSequenceNumber = signing.NextRequestSequenceNumber(sequenceNumber)

	return signing.ResponseSequenceNumber(sequenceNumber), true
}

// verifyIncoming validates the signature of a received response against the given
// expected sequence number when signing is active. It is a no-op when signing is
// inactive.
func (c *Client) verifyIncoming(raw []byte, responseSequenceNumber uint32) error {
	if c.Connection == nil || !c.Connection.IsSigningActive {
		return nil
	}
	if !signing.Verify(c.Connection.SigningSessionKey, raw, responseSequenceNumber) {
		return fmt.Errorf("invalid SMB message signature on response (expected sequence %d)", responseSequenceNumber)
	}
	return nil
}
