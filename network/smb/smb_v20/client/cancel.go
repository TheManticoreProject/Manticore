package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// Cancel requests cancellation of the operation this connection is waiting on —
// for example a blocked ChangeNotify. When the server has answered it with an
// interim STATUS_PENDING, Cancel sends an SMB2 CANCEL in the ASYNC form carrying
// its AsyncId; before that, it sends the synchronous form carrying the request's
// MessageId, TreeId and SessionId (MS-SMB2 3.2.4.24). CANCEL itself has no
// response; the cancelled operation instead completes with STATUS_CANCELLED,
// which its blocked caller observes.
//
// Cancel is intended to be called from a different goroutine than the one blocked
// in the operation. It returns an error if no operation is outstanding.
// Wire: SMB2 CANCEL.
func (c *Client) Cancel() error {
	// Both records are read under one lock, so the async state and the
	// synchronous record always describe the same request.
	c.asyncMu.Lock()
	messageId, asyncId, async := c.pendingMessageId, c.pendingAsyncId, c.hasPendingAsync
	treeId, sessionId, outstanding := c.outstandingTreeId, c.outstandingSessionId, c.hasOutstanding
	if !async {
		messageId = c.outstandingMessageId
	}
	c.asyncMu.Unlock()
	if !async && !outstanding {
		return fmt.Errorf("no operation is outstanding to cancel")
	}
	if !c.Transport.IsConnected() {
		return fmt.Errorf("cancel: transport is not connected")
	}

	// CANCEL reuses the target request's MessageId (it is matched to the
	// outstanding request, not allocated a new sequence number). Once an
	// interim response has given the request an AsyncId, it carries that in the
	// ASYNC header form; before, it names the request in the synchronous form.
	msg := message.NewMessage()
	msg.Header.MessageId = types.UINT64(messageId)
	msg.Header.Credit = 1
	if async {
		msg.Header.SetFlags(flags.SMB2_FLAGS_ASYNC_COMMAND)
		msg.Header.SetAsyncId(types.UINT64(asyncId))
		if c.Session != nil {
			msg.Header.SessionId = c.Session.SessionId
		}
	} else {
		// The synchronous form names the request by MessageId within the
		// same session and tree.
		msg.Header.TreeId = types.ULONG(treeId)
		msg.Header.SessionId = types.UINT64(sessionId)
	}
	msg.SetCommand(commands.NewCancelRequest())

	marshalled, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal cancel: %w", err)
	}

	// When the session encrypts data, the CANCEL is wrapped in an SMB2
	// TRANSFORM_HEADER instead of being signed — the AEAD tag supersedes the
	// per-message signature (MS-SMB2 3.2.4.24). Otherwise sign in place when
	// signing is active.
	encrypt := c.Session != nil && c.Session.EncryptData
	if !encrypt && c.Session != nil && c.Session.SigningActive {
		security.Sign(c.Connection.Dialect, c.Connection.SigningAlgorithmId, c.Session.SigningKey, marshalled)
	}

	wire := marshalled
	if encrypt {
		if wire, err = c.encryptMessage(marshalled); err != nil {
			return fmt.Errorf("failed to encrypt cancel: %w", err)
		}
	}

	if _, err := c.Transport.Send(wire); err != nil {
		return fmt.Errorf("failed to send cancel: %w", err)
	}
	return nil
}
