package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
)

// Negotiate initiates the SMB protocol negotiation with the server.
//
// This function performs the SMB_COM_NEGOTIATE exchange, which is the first step
// in establishing an SMB session. It sends a list of dialects supported by the client
// and receives the server's preferred dialect along with server capabilities.
//
// The negotiation process:
// 1. Creates and sends an SMB_COM_NEGOTIATE_REQUEST message
// 2. Receives the SMB_COM_NEGOTIATE_RESPONSE from the server
// 3. Validates the response command type
// 4. Processes server capabilities and configuration
//
// Returns:
//   - nil if negotiation is successful
//   - An error if any step in the negotiation process fails (connection issues,
//     message creation/marshalling errors, transport errors, or unexpected responses)
func (c *Client) Negotiate() error {
	if !c.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}

	requestMsg := message.NewMessage()

	requestMsg.Header.SetFlags(flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_CASE_INSENSITIVE)
	requestMsg.Header.SetFlags2(flags2.FLAGS2_UNICODE | flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_EXTENDED_SECURITY | flags2.FLAGS2_LONG_NAMES_ALLOWED)

	negotiateCmd := commands.NewNegotiateRequest()
	negotiateCmd.Dialects.AddDialect(dialects.DIALECT_NT_LM_0_12)

	requestMsg.AddCommand(negotiateCmd)

	marshalledMessage, err := requestMsg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal negotiate message: %v", err)
	}

	// Send the message
	_, err = c.Transport.Send(marshalledMessage)
	if err != nil {
		return fmt.Errorf("failed to send negotiate message: %v", err)
	}

	// Receive the response
	rawResponseMessage, err := c.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive response message: %v", err)
	}

	responseMsg := message.NewMessage()
	responseMsg.AddCommand(negotiateCmd)
	err = responseMsg.Unmarshal(rawResponseMessage)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response message: %v", err)
	}

	if responseMsg.Header.Command != codes.SMB_COM_NEGOTIATE {
		return fmt.Errorf("unexpected response command: %d", responseMsg.Header.Command)
	}

	negotiateResponse := responseMsg.Command.(*commands.NegotiateResponse)

	selectedDialect, err := negotiateResponse.GetSelectedDialect(negotiateCmd.Dialects)
	if err != nil {
		return fmt.Errorf("failed to get selected dialect: %v", err)
	}

	c.Connection.SelectedDialect = selectedDialect

	c.Connection.Server.Capabilities = negotiateResponse.Capabilities
	c.Connection.Server.SessionKey = negotiateResponse.SessionKey
	c.Connection.Server.SystemTime = negotiateResponse.SystemTime
	c.Connection.Server.TimeZone = negotiateResponse.ServerTimeZone
	c.Connection.Server.MaxBufferSize = negotiateResponse.MaxBufferSize
	c.Connection.MaxMpxCount = negotiateResponse.MaxMpxCount

	c.Connection.Server.DomainName = string(negotiateResponse.DomainName)
	c.Connection.Server.Name = string(negotiateResponse.ServerName)
	c.Connection.Server.SecurityMode = negotiateResponse.SecurityMode
	c.Connection.Server.ServerGUID = negotiateResponse.ServerGUID

	return nil
}
