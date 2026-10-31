package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// OplockBreak is a server-initiated oplock break: the open whose oplock is being
// broken and the level the server is breaking it down to (one of
// commands.SMB2_OPLOCK_LEVEL_*).
type OplockBreak struct {
	FileId   types.SMB2_FILEID
	NewLevel uint8
}

// WaitOplockBreak reads the next unsolicited SMB2 OPLOCK_BREAK notification from
// the server. The server sends one when another client opens a file on which this
// client holds an oplock (granted via CreateFileWithOplock); the caller should
// reply with AcknowledgeOplockBreak. WaitOplockBreak blocks until a notification
// arrives and is therefore typically run on a dedicated goroutine; it must not run
// concurrently with another operation that reads the connection.
func (c *Client) WaitOplockBreak() (*OplockBreak, error) {
	if !c.Transport.IsConnected() {
		return nil, fmt.Errorf("transport is not connected")
	}

	raw, wasEncrypted, err := c.receiveBreak()
	if err != nil {
		return nil, fmt.Errorf("failed to receive oplock break: %w", err)
	}

	if c.Session != nil && c.Session.EncryptData && !wasEncrypted {
		return nil, fmt.Errorf("oplock break: session requires encryption but notification was not encrypted")
	}

	msg := message.NewMessage()
	if _, err := msg.Header.Unmarshal(raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oplock break header: %w", err)
	}
	if !msg.Header.HasValidProtocolId() {
		return nil, fmt.Errorf("oplock break is not an SMB2 message (ProtocolId % x)", msg.Header.ProtocolId)
	}
	// The notification carries the reserved MessageId and is sent unsigned
	// (MS-SMB2 3.3.4.6), so it is exempt from verification (MS-SMB2 3.2.5.1.3).
	if !wasEncrypted && c.Session != nil && c.Session.SigningActive && signatureRequired(msg) {
		if !security.Verify(c.Connection.Dialect, c.Connection.SigningAlgorithmId, c.Session.SigningKey, raw) {
			return nil, fmt.Errorf("oplock break failed SMB2 signature verification")
		}
	}
	if msg.Header.Command != codes.SMB2_OPLOCK_BREAK {
		return nil, fmt.Errorf("expected an OPLOCK_BREAK notification, got command 0x%04x", uint16(msg.Header.Command))
	}
	if _, err := msg.Unmarshal(raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oplock break: %w", err)
	}
	notify, ok := msg.Command.(*commands.OplockBreakResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected oplock break command: %T", msg.Command)
	}
	return &OplockBreak{FileId: notify.FileId, NewLevel: uint8(notify.OplockLevel)}, nil
}

// AcknowledgeOplockBreak replies to an OPLOCK_BREAK notification, accepting the
// reduced oplock level on the given open, and waits for the server's response.
// Wire: SMB2 OPLOCK_BREAK Acknowledgment.
func (c *Client) AcknowledgeOplockBreak(fileId types.SMB2_FILEID, oplockLevel uint8) error {
	if c.Session == nil || c.Session.TreeId == 0 {
		return fmt.Errorf("no tree connect established")
	}

	req := commands.NewOplockBreakRequest()
	req.OplockLevel = types.UCHAR(oplockLevel)
	req.FileId = fileId

	response, err := c.sendReceive(c.newRequest(req), "OplockBreakAck")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("oplock break acknowledgment failed: %s", formatNTStatus(status))
	}
	return nil
}

// maxQueuedBreaks bounds the break notifications held for a later
// WaitOplockBreak / WaitBreakNotification. When it is reached the oldest is
// dropped, so a server that floods breaks cannot grow the queue without bound.
const maxQueuedBreaks = 64

// queuedBreak is an OPLOCK_BREAK notification set aside by a request that
// received it while waiting for its own response: the plaintext frame and
// whether it arrived wrapped in a TRANSFORM_HEADER.
type queuedBreak struct {
	raw       []byte
	encrypted bool
}

// queueBreak sets aside a break notification received while waiting for a
// response. raw is the plaintext frame (already decrypted when it arrived in a
// TRANSFORM_HEADER); it is copied, since callers reuse their buffers.
func (c *Client) queueBreak(raw []byte, encrypted bool) {
	c.breakMu.Lock()
	defer c.breakMu.Unlock()
	if len(c.pendingBreaks) >= maxQueuedBreaks {
		c.pendingBreaks = c.pendingBreaks[1:]
	}
	c.pendingBreaks = append(c.pendingBreaks, queuedBreak{raw: append([]byte(nil), raw...), encrypted: encrypted})
}

// receiveBreak returns the next break notification frame in plaintext and
// whether it was encrypted on the wire. A break queued by an earlier request is
// returned first; otherwise the next frame is read from the connection and, when
// per-session encryption is active and the server wrapped it in an SMB2
// TRANSFORM_HEADER (MS-SMB2 3.2.5.19), decrypted.
func (c *Client) receiveBreak() ([]byte, bool, error) {
	c.breakMu.Lock()
	if len(c.pendingBreaks) > 0 {
		next := c.pendingBreaks[0]
		c.pendingBreaks = c.pendingBreaks[1:]
		c.breakMu.Unlock()
		return next.raw, next.encrypted, nil
	}
	c.breakMu.Unlock()

	raw, err := c.Transport.Receive()
	if err != nil {
		return nil, false, err
	}
	if security.IsTransformHeader(raw) {
		plaintext, err := c.decryptMessage(raw)
		if err != nil {
			return nil, false, err
		}
		return plaintext, true, nil
	}
	return raw, false, nil
}
