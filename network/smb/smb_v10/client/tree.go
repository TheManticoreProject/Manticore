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
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	requestMsg := message.NewMessage()
	requestMsg.Header.Command = codes.SMB_COM_TREE_CONNECT_ANDX
	requestMsg.Header.Flags = 0x0000
	requestMsg.Header.Flags2 = 0x0000
	requestMsg.Header.SetPID(requestMsg.Header.GetPID())
	requestMsg.Header.MID = c.Connection.MaxMpxCount
	requestMsg.Header.TID = 65535
	requestMsg.Header.UID = c.Session.SessionUID

	treeConnectCmd := commands.NewTreeConnectAndxRequest()

	treeConnectCmd.Password = []types.UCHAR{}
	treeConnectCmd.PasswordLength = types.USHORT(0x0000)

	uncPath := "\\\\" + c.Connection.Server.Host.String() + "\\" + shareName + "\x00"
	treeConnectCmd.Path = []types.UCHAR(uncPath)

	treeConnectCmd.Service = []types.UCHAR("?????" + "\x00")

	treeConnectCmd.Flags = 0x0000
	treeConnectCmd.Pad = []types.UCHAR{}

	requestMsg.AddCommand(treeConnectCmd)

	marshalledMessage, err := requestMsg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal tree connect message: %v", err)
	}

	_, err = c.Transport.Send(marshalledMessage)
	if err != nil {
		return fmt.Errorf("failed to send tree connect message: %v", err)
	}

	rawResponseMessage, err := c.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive tree connect message: %v", err)
	}

	responseMsg := message.NewMessage()
	err = responseMsg.Unmarshal(rawResponseMessage)
	if err != nil {
		return fmt.Errorf("failed to unmarshal tree connect message: %v", err)
	}

	if responseMsg.Header.Command != codes.SMB_COM_TREE_CONNECT_ANDX {
		return fmt.Errorf("unexpected response command: 0x%02x", responseMsg.Header.Command)
	}

	treeConnectResponse := responseMsg.Command.(*commands.TreeConnectAndxResponse)

	if responseMsg.Header.Status != 0x00 {
		return fmt.Errorf("tree connect failed: 0x%08x", responseMsg.Header.Status)
	}

	if c.Connection.TreeConnectTable == nil {
		c.Connection.TreeConnectTable = make(map[uint16]interface{})
	}
	c.Connection.TreeConnectTable[responseMsg.Header.TID] = treeConnectResponse

	// Select this tree as the current one for subsequent commands (file I/O, etc.).
	if c.Session != nil {
		c.Session.TreeID = responseMsg.Header.TID
	}

	return nil
}
