package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// TreeConnect connects to a share on the server and selects it for subsequent
// file operations. The server returns the assigned TreeId in the response
// header; it is recorded on the session and in the connection's tree-connect
// table.
func (c *Client) TreeConnect(shareName string) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	req := commands.NewTreeConnectRequest()
	// SMB2 tree-connect paths are the full UNC form \\server\share.
	req.Path = fmt.Sprintf("\\\\%s\\%s", c.Connection.Server.Host.String(), shareName)

	msg := c.newRequest(req)
	// TreeId MUST be 0 in a TREE_CONNECT request.
	msg.Header.TreeId = 0

	response, err := c.sendReceive(msg, "TreeConnect")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("tree connect to %q failed: %s", shareName, formatNTStatus(status))
	}

	treeConnectResponse, ok := response.Command.(*commands.TreeConnectResponse)
	if !ok {
		return fmt.Errorf("unexpected tree connect response command: %T", response.Command)
	}

	// The server assigns the TreeId in the response header.
	treeId := response.Header.TreeId
	c.Session.TreeId = treeId

	// An SMB 3.x share may require encryption; when the server sets the
	// SMB2_SHAREFLAG_ENCRYPT_DATA flag, encrypt all subsequent traffic on this
	// session.
	if isSMB3Dialect(c.Connection.Dialect) &&
		treeConnectResponse.ShareFlags&commands.SMB2_SHAREFLAG_ENCRYPT_DATA != 0 &&
		len(c.Session.EncryptionKey) > 0 {
		c.Session.EncryptData = true
	}

	tc := &TreeConnect{
		Connection: c.Connection,
		Session:    c.Session,
		ShareName:  shareName,
		TreeId:     treeId,
		ShareType:  treeConnectResponse.ShareType,
	}
	if c.Connection.TreeConnectTable == nil {
		c.Connection.TreeConnectTable = make(map[uint32]*TreeConnect)
	}
	c.Connection.TreeConnectTable[treeId] = tc

	return nil
}
