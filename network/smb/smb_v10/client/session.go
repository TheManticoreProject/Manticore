package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// Session represents an established session between the client and server
type Session struct {
	Connection      *Client     // The SMB connection associated with this session
	SessionKey      []byte      // The cryptographic session key associated with this session
	SessionUID      uint16      // The 2-byte UID for this session
	UserCredentials interface{} // Opaque implementation-specific entity that identifies the credentials
}

func (c *Client) SessionSetup() error {
	if !c.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}

	request_msg := message.NewMessage()
	session_setup_cmd := commands.NewSessionSetupAndxRequest()

	session_setup_cmd.MaxBufferSize = uint16(c.Connection.Server.MaxBufferSize)
	session_setup_cmd.MaxMpxCount = c.Connection.MaxMpxCount
	session_setup_cmd.VcNumber = 0x0000
	session_setup_cmd.SessionKey = c.Connection.Server.SessionKey
	session_setup_cmd.OEMPasswordLen = 0x0000
	session_setup_cmd.UnicodePasswordLen = 0x0000
	session_setup_cmd.Reserved = 0x00000000
	session_setup_cmd.Capabilities = c.Connection.Server.Capabilities

	// Data section
	session_setup_cmd.OEMPassword = []types.UCHAR("OEMPassword")
	session_setup_cmd.OEMPasswordLen = uint16(len(session_setup_cmd.OEMPassword))
	session_setup_cmd.UnicodePassword = []types.UCHAR(utf16.EncodeUTF16LE("UnicodePassword"))
	session_setup_cmd.UnicodePasswordLen = uint16(len(session_setup_cmd.UnicodePassword))
	// session_setup_cmd.Pad = []types.UCHAR{0x00, 0x00, 0x00, 0x00}
	session_setup_cmd.AccountName.SetString("AccountName")
	session_setup_cmd.PrimaryDomain.SetString("DomainName")
	session_setup_cmd.NativeOS.SetString("")
	session_setup_cmd.NativeLanMan.SetString("")

	request_msg.AddCommand(session_setup_cmd)

	marshalled_message, err := request_msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal negotiate message: %v", err)
	}

	fmt.Printf("marshalled_message: %v\n", marshalled_message)

	// Send the message
	_, err = c.Transport.Send(marshalled_message)
	if err != nil {
		return fmt.Errorf("failed to send negotiate message: %v", err)
	}

	// Receive the response
	raw_response_message, err := c.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive response message: %v", err)
	}

	response_msg := message.NewMessage()
	err = response_msg.Unmarshal(raw_response_message)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response message: %v", err)
	}

	if response_msg.Header.Command != codes.SMB_COM_SESSION_SETUP_ANDX {
		return fmt.Errorf("unexpected response command: %d", response_msg.Header.Command)
	}

	_ = response_msg.Command.(*commands.SessionSetupAndxResponse)

	// TODO

	return nil
}
