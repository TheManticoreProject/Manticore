package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
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

	raw, err := c.Transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive oplock break: %w", err)
	}

	msg := message.NewMessage()
	if _, err := msg.Header.Unmarshal(raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oplock break header: %w", err)
	}
	if !msg.Header.HasValidProtocolId() {
		return nil, fmt.Errorf("oplock break is not an SMB2 message (ProtocolId % x)", msg.Header.ProtocolId)
	}
	if c.Session != nil && c.Session.SigningActive && msg.Header.Flags.IsSigned() {
		if !verifySignature(c.Session.SigningKey, raw) {
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
