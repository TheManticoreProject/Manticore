package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// Flush requests that the server flush all buffered data for the open file
// referenced by fid to stable storage. Passing FID 0xFFFF flushes every file the
// client has open on the connection.
//
// Wire: SMB_COM_FLUSH request / response.
func (c *Client) Flush(fid FID) error {
	if c.Session == nil {
		return fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_FLUSH)

	cmd := commands.NewFlushRequest()
	cmd.FID = types.USHORT(fid)

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "Flush")
	if err != nil {
		return err
	}

	if response.Header.Status != 0x00000000 {
		return fmt.Errorf("Flush failed: 0x%08x", response.Header.Status)
	}

	return nil
}

// Echo sends an SMB_COM_ECHO request carrying data and returns the data echoed
// back by the server. It is a round-trip/keepalive check on the connection.
//
// Wire: SMB_COM_ECHO request / response (EchoCount 1).
func (c *Client) Echo(data []byte) ([]byte, error) {
	if c.Session == nil {
		return nil, fmt.Errorf("no session established")
	}

	msg := c.newFileIOMessage(codes.SMB_COM_ECHO)

	cmd := commands.NewEchoRequest()
	// Request a single echo response carrying our data back.
	cmd.EchoCount = types.USHORT(1)
	cmd.Data = []types.UCHAR(data)

	msg.AddCommand(cmd)

	response, _, err := c.sendReceive(msg, "Echo")
	if err != nil {
		return nil, err
	}

	if response.Header.Status != 0x00000000 {
		return nil, fmt.Errorf("Echo failed: 0x%08x", response.Header.Status)
	}

	echoResponse, ok := response.Command.(*commands.EchoResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response command: 0x%02x", response.Header.Command)
	}

	return []byte(echoResponse.Data), nil
}
