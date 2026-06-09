package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// Flush issues an SMB2 FLUSH, asking the server to commit any buffered data for
// the open file to stable storage before returning. Wire: SMB2 FLUSH.
func (c *Client) Flush(fileId types.SMB2_FILEID) error {
	if c.Session == nil || c.Session.TreeId == 0 {
		return fmt.Errorf("no tree connect established")
	}

	req := commands.NewFlushRequest()
	req.FileId = fileId

	response, err := c.sendReceive(c.newRequest(req), "Flush")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("flush failed: %s", formatNTStatus(status))
	}
	return nil
}
