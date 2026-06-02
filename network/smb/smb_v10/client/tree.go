package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TreeConnect represents an established tree connect between the client and share on the server
type TreeConnect struct {
	Connection *Connection // The SMB connection associated with this tree connect
	ShareName  string      // The share name corresponding to this tree connect
	TreeID     uint16      // The TreeID (TID) that identifies this tree connect
	Session    *Session    // A reference to the session on which this tree connect was established
	IsDfsShare bool        // A Boolean that, if set, indicates that the tree connect was established to a DFS share
}

func (c *Client) TreeConnect(shareName string) error {
	request_msg := message.NewMessage()
	request_msg.Header.Command = codes.SMB_COM_TREE_CONNECT_ANDX
	request_msg.Header.Flags = 0x0000
	request_msg.Header.Flags2 = 0x0000
	request_msg.Header.SetPID(request_msg.Header.GetPID())
	request_msg.Header.MID = c.Connection.MaxMpxCount
	request_msg.Header.TID = 65535
	request_msg.Header.UID = c.Session.SessionUID

	tree_connect_cmd := commands.NewTreeConnectAndxRequest()

	tree_connect_cmd.Password = []types.UCHAR{}
	tree_connect_cmd.PasswordLength = types.USHORT(0x0000)

	uncPath := "\\\\" + c.Connection.Server.Host.String() + "\\" + shareName + "\x00"
	tree_connect_cmd.Path = []types.UCHAR(uncPath)

	tree_connect_cmd.Service = []types.UCHAR("?????" + "\x00")

	tree_connect_cmd.Flags = 0x0000
	tree_connect_cmd.Pad = []types.UCHAR{}

	request_msg.AddCommand(tree_connect_cmd)

	marshalled_message, err := request_msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal tree connect message: %v", err)
	}

	_, err = c.Transport.Send(marshalled_message)
	if err != nil {
		return fmt.Errorf("failed to send tree connect message: %v", err)
	}

	raw_response_message, err := c.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive tree connect message: %v", err)
	}

	response_msg := message.NewMessage()
	err = response_msg.Unmarshal(raw_response_message)
	if err != nil {
		return fmt.Errorf("failed to unmarshal tree connect message: %v", err)
	}

	if response_msg.Header.Command != codes.SMB_COM_TREE_CONNECT_ANDX {
		return fmt.Errorf("unexpected response command: 0x%02x", response_msg.Header.Command)
	}

	tree_connect_response_cmd := response_msg.Command.(*commands.TreeConnectAndxResponse)

	if response_msg.Header.Status != 0x00 {
		return fmt.Errorf("tree connect failed: 0x%08x", response_msg.Header.Status)
	}

	if c.Connection.TreeConnectTable == nil {
		c.Connection.TreeConnectTable = make(map[uint16]interface{})
	}
	c.Connection.TreeConnectTable[response_msg.Header.TID] = tree_connect_response_cmd

	// Select this tree as the current one for subsequent commands (file I/O, etc.).
	if c.Session != nil {
		c.Session.TreeID = response_msg.Header.TID
	}

	return nil
}
