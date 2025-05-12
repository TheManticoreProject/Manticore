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

	request_msg := message.NewMessage()

	request_msg.Header.SetFlags(flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_CASE_INSENSITIVE)
	request_msg.Header.SetFlags2(flags2.FLAGS2_UNICODE | flags2.FLAGS2_LONG_NAMES_ALLOWED | flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_SECURITY_SIGNATURE)

	negotiate_cmd := commands.NewNegotiateRequest()
	negotiate_cmd.Dialects.AddDialect(dialects.DIALECT_NT_LM_0_12)

	request_msg.AddCommand(negotiate_cmd)

	marshalled_message, err := request_msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal negotiate message: %v", err)
	}

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
	response_msg.AddCommand(negotiate_cmd)
	err = response_msg.Unmarshal(raw_response_message)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response message: %v", err)
	}

	if response_msg.Header.Command != codes.SMB_COM_NEGOTIATE {
		return fmt.Errorf("unexpected response command: %d", response_msg.Header.Command)
	}

	negotiate_response := response_msg.Command.(*commands.NegotiateResponse)

	selected_dialect, err := negotiate_response.GetSelectedDialect(negotiate_cmd.Dialects)
	if err != nil {
		return fmt.Errorf("failed to get selected dialect: %v", err)
	}

	c.Connection.SelectedDialect = selected_dialect

	c.Connection.Server.Capabilities = negotiate_response.Capabilities
	c.Connection.Server.SessionKey = negotiate_response.SessionKey
	c.Connection.Server.SystemTime = negotiate_response.SystemTime
	c.Connection.Server.TimeZone = negotiate_response.ServerTimeZone
	c.Connection.Server.MaxBufferSize = negotiate_response.MaxBufferSize
	c.Connection.MaxMpxCount = negotiate_response.MaxMpxCount

	c.Connection.Server.DomainName = string(negotiate_response.DomainName)
	c.Connection.Server.Name = string(negotiate_response.ServerName)
	c.Connection.Server.SecurityMode = negotiate_response.SecurityMode

	return nil
}
