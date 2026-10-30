package client_test

import (
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/securitymode"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// marshalWithoutReplyFlag builds a server "reply" whose header lacks
// SMB_FLAGS_REPLY, so message.Unmarshal decodes its body as the Request type
// for the command code rather than the Response type.
func marshalWithoutReplyFlag(t *testing.T, cmd command_interface.CommandInterface) []byte {
	t.Helper()
	m := message.NewMessage()
	m.AddCommand(cmd)
	raw, err := m.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal scripted reply: %v", err)
	}
	return raw
}

// TestClientRejectsReplyWithoutReplyFlag verifies that Negotiate,
// SessionSetup and TreeConnect return an error, instead of panicking on a
// failed type assertion, when the server's reply does not carry
// SMB_FLAGS_REPLY and is therefore decoded as a request. CloseFile covers the
// shared sendReceive path, whose callers must not accept such a message as a
// successful reply.
func TestClientRejectsReplyWithoutReplyFlag(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T) error
	}{
		{
			name: "Negotiate",
			run: func(t *testing.T) error {
				tr := &scriptedTransport{response: marshalWithoutReplyFlag(t, commands.NewNegotiateRequest())}
				c := &client.Client{Transport: tr, Connection: &client.Connection{Server: &client.Server{}}}
				return c.Negotiate()
			},
		},
		{
			name: "SessionSetup",
			run: func(t *testing.T) error {
				tr := &scriptedTransport{response: marshalWithoutReplyFlag(t, commands.NewSessionSetupAndxRequest())}
				c := &client.Client{Transport: tr, Connection: &client.Connection{Server: &client.Server{SecurityMode: securitymode.SecurityMode(0x03)}}}
				return c.SessionSetup(&credentials.Credentials{Username: "u", Password: "p"})
			},
		},
		{
			name: "TreeConnect",
			run: func(t *testing.T) error {
				tr := &scriptedTransport{response: marshalWithoutReplyFlag(t, commands.NewTreeConnectAndxRequest())}
				c := &client.Client{Transport: tr, Connection: &client.Connection{Server: &client.Server{Host: net.IPv4(1, 2, 3, 4)}}}
				c.Session = &client.Session{SessionUID: 1}
				return c.TreeConnect("IPC$")
			},
		},
		{
			name: "CloseFile",
			run: func(t *testing.T) error {
				tr := &scriptedTransport{response: marshalWithoutReplyFlag(t, commands.NewCloseRequest())}
				c := &client.Client{Transport: tr, Connection: &client.Connection{Server: &client.Server{}}}
				c.Session = &client.Session{SessionUID: 1, TreeID: 1}
				return c.CloseFile(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s panicked on a reply lacking SMB_FLAGS_REPLY: %v", tt.name, r)
					}
				}()
				err = tt.run(t)
			}()
			if err == nil {
				t.Fatalf("%s accepted a reply lacking SMB_FLAGS_REPLY; want an error", tt.name)
			}
		})
	}
}
