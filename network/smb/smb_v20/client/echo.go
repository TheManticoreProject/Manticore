package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// Echo sends an SMB2 ECHO and waits for the reply. ECHO carries no data; it is a
// keepalive that confirms the connection (and, if established, the session) is
// still alive. Wire: SMB2 ECHO.
func (c *Client) Echo() error {
	response, err := c.sendReceive(c.newRequest(commands.NewEchoRequest()), "Echo")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("echo failed: %s", formatNTStatus(status))
	}
	return nil
}
