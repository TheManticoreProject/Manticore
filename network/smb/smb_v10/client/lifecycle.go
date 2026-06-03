package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
)

// TreeDisconnect releases the current tree connection (TID) on the server.
//
// Wire: SMB_COM_TREE_DISCONNECT request / response.
func (c *Client) TreeDisconnect() error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_TREE_DISCONNECT)
	cmd := commands.NewTreeDisconnectRequest()
	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "TreeDisconnect")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("TreeDisconnect failed: 0x%08x", response.Header.Status)
	}

	// Remove the TID from the local tree-connect table.
	tid := c.Session.TreeID
	if c.Connection.TreeConnectTable != nil {
		delete(c.Connection.TreeConnectTable, tid)
	}
	c.Session.TreeID = 0

	return nil
}

// Logoff releases the current session (UID) on the server.
//
// Wire: SMB_COM_LOGOFF_ANDX request / response.
func (c *Client) Logoff() error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_LOGOFF_ANDX)
	cmd := commands.NewLogoffAndxRequest()
	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "LogoffAndx")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("LogoffAndx failed: 0x%08x", response.Header.Status)
	}

	// Remove the session from the local session table.
	uid := c.Session.SessionUID
	if c.Connection.SessionTable != nil {
		delete(c.Connection.SessionTable, uid)
	}
	c.Session = nil

	return nil
}

// Disconnect closes the underlying transport connection to the server.
// It does not send any SMB commands; call TreeDisconnect and Logoff first
// for a clean teardown.
func (c *Client) Disconnect() error {
	if c.Transport == nil {
		return nil
	}
	return c.Transport.Close()
}
