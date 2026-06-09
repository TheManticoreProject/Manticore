package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// TreeDisconnect releases the currently selected tree connect (TreeId).
//
// Wire: SMB2 TREE_DISCONNECT request / response.
func (c *Client) TreeDisconnect() error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}
	if c.Session.TreeId == 0 {
		return fmt.Errorf("no tree connect established")
	}

	msg := c.newRequest(commands.NewTreeDisconnectRequest())

	response, err := c.sendReceive(msg, "TreeDisconnect")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("tree disconnect failed: %s", formatNTStatus(status))
	}

	treeId := c.Session.TreeId
	if c.Connection.TreeConnectTable != nil {
		delete(c.Connection.TreeConnectTable, treeId)
	}
	c.Session.TreeId = 0

	return nil
}

// Logoff releases the current session (SessionId).
//
// Wire: SMB2 LOGOFF request / response.
func (c *Client) Logoff() error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newRequest(commands.NewLogoffRequest())

	response, err := c.sendReceive(msg, "Logoff")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("logoff failed: %s", formatNTStatus(status))
	}

	sessionId := c.Session.SessionId
	if c.Connection.SessionTable != nil {
		delete(c.Connection.SessionTable, sessionId)
	}
	c.Session = nil

	return nil
}
