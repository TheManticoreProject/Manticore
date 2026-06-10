package client_test

import (
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
)

// cannedTransport is a transport.Transport stub that reports itself connected,
// accepts any Send, and returns a fixed response from Receive.
type cannedTransport struct {
	response []byte
}

func (m *cannedTransport) Connect(ipaddr net.IP, port int) error { return nil }
func (m *cannedTransport) Close() error                          { return nil }
func (m *cannedTransport) Send(data []byte) (int, error)         { return len(data), nil }
func (m *cannedTransport) Receive() ([]byte, error)              { return m.response, nil }
func (m *cannedTransport) IsConnected() bool                     { return true }
func (m *cannedTransport) SetTimeout(time.Duration)              {}

// TestNegotiateRejectsErrorStatus verifies that Negotiate returns an error when
// the server's NEGOTIATE response carries a non-zero Header.Status, rather than
// treating the error response as a successful negotiation.
func TestNegotiateRejectsErrorStatus(t *testing.T) {
	// Build a NEGOTIATE response message whose header reports an error status but
	// whose command body is otherwise well-formed (DialectIndex 0 selects a valid
	// dialect, so only the status check should cause a failure).
	respMsg := message.NewMessage()
	respMsg.AddCommand(commands.NewNegotiateResponse())
	respMsg.Header.Command = codes.SMB_COM_NEGOTIATE
	// Mark the message as a reply so the client unmarshals it as a NegotiateResponse.
	respMsg.Header.SetFlags(flags.FLAGS_REPLY)
	respMsg.Header.Status = 0xC0000022 // STATUS_ACCESS_DENIED

	raw, err := respMsg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal canned response: %v", err)
	}

	c := &client.Client{
		Transport:  &cannedTransport{response: raw},
		Connection: &client.Connection{Server: &client.Server{}},
	}

	if err := c.Negotiate(); err == nil {
		t.Fatal("expected Negotiate to fail on an error-status response, got nil")
	}
}
