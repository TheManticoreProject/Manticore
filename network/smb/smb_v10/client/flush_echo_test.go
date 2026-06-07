package client_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// replyBytes builds a marshalled reply message carrying cmd with the given status.
func replyBytes(t *testing.T, status uint32, cmd command_interface.CommandInterface) []byte {
	t.Helper()
	msg := message.NewMessage()
	msg.Header.SetFlags(flags.FLAGS_REPLY)
	msg.Header.Status = status
	msg.AddCommand(cmd)
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal reply: %v", err)
	}
	return raw
}

// sessionedClient returns a client with a session and a scripted transport that
// replies with response.
func sessionedClient(response []byte) *client.Client {
	tr := &scriptedTransport{response: response}
	return &client.Client{
		Transport:  tr,
		Connection: &client.Connection{Server: &client.Server{}},
		Session:    &client.Session{},
	}
}

func TestFlushSuccess(t *testing.T) {
	c := sessionedClient(replyBytes(t, 0x00000000, commands.NewFlushResponse()))
	if err := c.Flush(0xFFFF); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}
}

func TestFlushErrorStatus(t *testing.T) {
	c := sessionedClient(replyBytes(t, 0xC0000008, commands.NewFlushResponse())) // STATUS_INVALID_HANDLE
	if err := c.Flush(1); err == nil {
		t.Fatal("expected an error on non-zero status, got nil")
	}
}

func TestFlushWithoutSession(t *testing.T) {
	c := &client.Client{Connection: &client.Connection{Server: &client.Server{}}}
	if err := c.Flush(1); err == nil {
		t.Fatal("expected an error when Flush is called without a session, got nil")
	}
}

func TestEchoRoundTrip(t *testing.T) {
	payload := []byte("keepalive-123")
	resp := commands.NewEchoResponse()
	resp.SequenceNumber = 1
	resp.Data = []types.UCHAR(payload)

	c := sessionedClient(replyBytes(t, 0x00000000, resp))

	got, err := c.Echo(payload)
	if err != nil {
		t.Fatalf("Echo returned error: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("Echo returned %q, want %q", got, payload)
	}
}

func TestEchoWithoutSession(t *testing.T) {
	c := &client.Client{Connection: &client.Connection{Server: &client.Server{}}}
	if _, err := c.Echo([]byte("x")); err == nil {
		t.Fatal("expected an error when Echo is called without a session, got nil")
	}
}
