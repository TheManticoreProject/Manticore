package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// sentSessionSetup drives SessionSetup against a canned response and returns the
// SESSION_SETUP_ANDX the client put on the wire.
func sentSessionSetup(t *testing.T, serverMaxBufferSize uint32, clientMaxBufferSize uint16) *commands.SessionSetupAndxRequest {
	t.Helper()

	resp := message.NewMessage()
	resp.Header.SetFlags(flags.FLAGS_REPLY)
	resp.Header.UID = 0x0901
	resp.AddCommand(commands.NewSessionSetupAndxResponse())
	raw, err := resp.Marshal()
	if err != nil {
		t.Fatalf("marshal canned response: %v", err)
	}

	tr := &capturingTransport{response: raw}
	c := &client.Client{
		Transport:     tr,
		Connection:    &client.Connection{Server: &client.Server{MaxBufferSize: serverMaxBufferSize}},
		MaxBufferSize: clientMaxBufferSize,
	}
	if err := c.SessionSetup(&credentials.Credentials{Username: "alice"}); err != nil {
		t.Fatalf("SessionSetup: %v", err)
	}

	reqMsg := message.NewMessage()
	if err := reqMsg.Unmarshal(tr.sent); err != nil {
		t.Fatalf("unmarshal sent request: %v", err)
	}
	req, ok := reqMsg.Command.(*commands.SessionSetupAndxRequest)
	if !ok {
		t.Fatalf("sent command is %T, want *SessionSetupAndxRequest", reqMsg.Command)
	}
	return req
}

// TestSessionSetupDeclaresClientMaxBufferSize guards the truncation bug: the
// SESSION_SETUP_ANDX MaxBufferSize field is 16 bits and describes the *client's*
// receive buffer ([MS-CIFS] 2.2.4.53.1). Narrowing the server's 32-bit advertised
// value into it discarded the high half, and a server advertising an exact multiple
// of 65536 made the client declare that it could receive nothing.
func TestSessionSetupDeclaresClientMaxBufferSize(t *testing.T) {
	tests := []struct {
		name             string
		serverAdvertised uint32
	}{
		{"server advertises 16644", 16644},
		{"server advertises 64 KiB", 65536},   // truncated to 0 before the fix
		{"server advertises 128 KiB", 131072}, // truncated to 0 before the fix
		{"server advertises 66000", 66000},    // truncated to 464 before the fix
		{"server advertises nothing", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := sentSessionSetup(t, tt.serverAdvertised, 0)
			if req.MaxBufferSize != uint16(client.DefaultMaxBufferSize) {
				t.Errorf("MaxBufferSize = %d, want the client default %d (server advertised %d)",
					req.MaxBufferSize, client.DefaultMaxBufferSize, tt.serverAdvertised)
			}
			if req.MaxBufferSize == 0 {
				t.Error("MaxBufferSize = 0: the client declared that it can receive no message at all")
			}
		})
	}
}

// TestSessionSetupHonoursConfiguredMaxBufferSize checks that a caller can declare
// its own buffer size.
func TestSessionSetupHonoursConfiguredMaxBufferSize(t *testing.T) {
	req := sentSessionSetup(t, 65536, 4356)
	if req.MaxBufferSize != 4356 {
		t.Errorf("MaxBufferSize = %d, want the configured 4356", req.MaxBufferSize)
	}
}
