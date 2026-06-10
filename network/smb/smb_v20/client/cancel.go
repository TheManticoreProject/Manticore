package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// Cancel requests cancellation of the operation currently awaiting an async
// (STATUS_PENDING) completion on this connection — for example a blocked
// ChangeNotify. It sends an SMB2 CANCEL carrying that request's MessageId and
// AsyncId. CANCEL itself has no response; the cancelled operation instead
// completes with STATUS_CANCELLED, which its blocked caller observes.
//
// Cancel is intended to be called from a different goroutine than the one blocked
// in the operation. It returns an error if no operation is currently pending.
// Wire: SMB2 CANCEL.
func (c *Client) Cancel() error {
	messageId, asyncId, ok := c.pendingAsync()
	if !ok {
		return fmt.Errorf("no async operation is pending to cancel")
	}
	if !c.Transport.IsConnected() {
		return fmt.Errorf("cancel: transport is not connected")
	}

	// CANCEL reuses the target request's MessageId (it is matched to the
	// outstanding request, not allocated a new sequence number) and carries the
	// AsyncId in the ASYNC header form.
	msg := message.NewMessage()
	msg.Header.MessageId = messageId
	msg.Header.Credit = 1
	msg.Header.SetFlags(flags.SMB2_FLAGS_ASYNC_COMMAND)
	msg.Header.SetAsyncId(types.UINT64(asyncId))
	if c.Session != nil {
		msg.Header.SessionId = c.Session.SessionId
	}
	msg.SetCommand(commands.NewCancelRequest())

	marshalled, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal cancel: %w", err)
	}
	if c.Session != nil && c.Session.SigningActive {
		signMessage(c.Session.SigningKey, marshalled)
	}
	if _, err := c.Transport.Send(marshalled); err != nil {
		return fmt.Errorf("failed to send cancel: %w", err)
	}
	return nil
}
