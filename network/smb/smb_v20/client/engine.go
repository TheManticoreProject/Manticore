package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
)

// newRequest builds an SMB2 request message with the header fields common to
// every command: the next 64-bit MessageId on the connection, a one-credit
// request (CreditCharge is 0 in the SMB 2.0.2 dialect), and — when a session is
// established — the current SessionId and TreeId. The command code is taken from
// the command when it is attached via SetCommand.
func (c *Client) newRequest(command command_interface.CommandInterface) *message.Message {
	msg := message.NewMessage()

	// Allocate the next MessageId for this connection.
	msg.Header.MessageId = c.Connection.MessageId
	c.Connection.MessageId++

	// SMB 2.0.2: CreditCharge MUST be 0; request a single credit per message.
	msg.Header.CreditCharge = 0
	msg.Header.Credit = 1

	if c.Session != nil {
		msg.Header.SessionId = c.Session.SessionId
		msg.Header.TreeId = c.Session.TreeId
	}

	msg.SetCommand(command)
	return msg
}

// sendReceive marshals and sends a request, then receives and parses the
// response. It records the credits the server grants (the response Credit field)
// and validates that the response carries the SMB2 protocol identifier.
func (c *Client) sendReceive(msg *message.Message, label string) (*message.Message, error) {
	if !c.Transport.IsConnected() {
		return nil, fmt.Errorf("%s: transport is not connected", label)
	}

	marshalled, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s: %w", label, err)
	}

	if _, err = c.Transport.Send(marshalled); err != nil {
		return nil, fmt.Errorf("failed to send %s: %w", label, err)
	}

	raw, err := c.Transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive %s response: %w", label, err)
	}

	response := message.NewMessage()
	if _, err = response.Unmarshal(raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s response: %w", label, err)
	}

	if !response.Header.HasValidProtocolId() {
		return nil, fmt.Errorf("%s response is not an SMB2 message (ProtocolId % x)", label, response.Header.ProtocolId)
	}

	// The server replenishes credits via the response Credit field.
	if response.Header.Credit > 0 {
		c.Connection.Credits = response.Header.Credit
	}

	return response, nil
}

// statusFromResponse returns the NT status code carried in a response header.
func statusFromResponse(response *message.Message) uint32 {
	return response.Header.Status
}
