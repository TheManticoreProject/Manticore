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

	// Sign the request in place when signing is active for the session.
	if c.Session != nil && c.Session.SigningActive {
		signMessage(c.Session.SigningKey, marshalled)
	}

	if _, err = c.Transport.Send(marshalled); err != nil {
		return nil, fmt.Errorf("failed to send %s: %w", label, err)
	}

	// Receive the response. The server may first return one or more interim
	// STATUS_PENDING responses (SMB2_FLAGS_ASYNC_COMMAND) for an operation it
	// cannot complete immediately (a blocking lock, CHANGE_NOTIFY, a long IOCTL);
	// each is verified and credited like any other response, then we keep reading
	// until the final response arrives.
	var raw []byte
	response := message.NewMessage()
	for {
		raw, err = c.Transport.Receive()
		if err != nil {
			return nil, fmt.Errorf("failed to receive %s response: %w", label, err)
		}

		response = message.NewMessage()
		if _, err = response.Header.Unmarshal(raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s response header: %w", label, err)
		}
		if !response.Header.HasValidProtocolId() {
			return nil, fmt.Errorf("%s response is not an SMB2 message (ProtocolId % x)", label, response.Header.ProtocolId)
		}

		// Skip server-initiated messages that can interleave with a synchronous
		// exchange — most notably an OPLOCK_BREAK notification, which the server
		// may send at any time once an oplock is granted and which carries the
		// reserved MessageId 0xFFFFFFFFFFFFFFFF. Such a message is not the
		// response to this request; consuming it as one would desynchronize the
		// request/response stream. Keep reading for the real response.
		if uint64(response.Header.MessageId) == unsolicitedMessageId {
			continue
		}

		// Verify the signature when signing is active and the server signed the
		// response (interim and final responses are both signed).
		if c.Session != nil && c.Session.SigningActive && response.Header.Flags.IsSigned() {
			if !verifySignature(c.Session.SigningKey, raw) {
				return nil, fmt.Errorf("%s response failed SMB2 signature verification", label)
			}
		}
		// The server replenishes credits via the response Credit field.
		if response.Header.Credit > 0 {
			c.Connection.Credits = response.Header.Credit
		}

		// A response must carry the MessageId of the request it answers; an
		// interim STATUS_PENDING response and the final response both echo it.
		// A mismatch means the stream is desynchronized.
		if response.Header.MessageId != msg.Header.MessageId {
			return nil, fmt.Errorf("%s response MessageId %d does not match request MessageId %d", label, uint64(response.Header.MessageId), uint64(msg.Header.MessageId))
		}

		if response.Header.Status != ntStatusPending {
			break
		}
		// Record this interim async operation so a concurrent Cancel can target it
		// (e.g. to interrupt a blocked CHANGE_NOTIFY), then wait for the final
		// response.
		c.setPendingAsync(msg.Header.MessageId, uint64(response.Header.GetAsyncId()))
	}
	c.clearPendingAsync()

	// The response command code must match the request's; otherwise the server
	// answered with a different command than was asked, indicating a corrupt or
	// desynchronized exchange.
	if response.Header.Command != msg.Header.Command {
		return nil, fmt.Errorf("%s response command 0x%04x does not match request command 0x%04x", label, uint16(response.Header.Command), uint16(msg.Header.Command))
	}

	// Decode the command body only for statuses that carry a real command response:
	// success, the session-setup continuation status, and STATUS_BUFFER_OVERFLOW (a
	// warning returned from READ / IOCTL pipe transceive with the partial data). An
	// SMB2 error response carries a fixed 9-byte SMB2 ERROR Response body
	// (MS-SMB2 2.2.2), not the command-specific response.
	status := response.Header.Status
	if status == 0x00000000 || status == ntStatusMoreProcessingRequired || status == ntStatusBufferOverflow {
		if _, err = response.Unmarshal(raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s response: %w", label, err)
		}
	}

	return response, nil
}

// unsolicitedMessageId is the reserved MessageId (0xFFFFFFFFFFFFFFFF) the server
// stamps on messages not sent in reply to a client request, such as an
// OPLOCK_BREAK notification (MS-SMB2 2.2.1.2).
const unsolicitedMessageId = 0xFFFFFFFFFFFFFFFF

// statusFromResponse returns the NT status code carried in a response header.
func statusFromResponse(response *message.Message) uint32 {
	return response.Header.Status
}
