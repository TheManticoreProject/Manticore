package client_test

import (
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
)

func sentFlags2(t *testing.T, raw []byte) flags2.Flags2 {
	t.Helper()
	request := message.NewMessage()
	if err := request.Unmarshal(raw); err != nil {
		t.Fatalf("unmarshal sent request: %v", err)
	}
	return request.Header.Flags2
}

func TestPostAuthenticationRequestsPreserveExtendedSecurity(t *testing.T) {
	for _, test := range []struct {
		name string
		caps capabilities.Capabilities
		want bool
	}{
		{"extended security negotiated", capabilities.CAP_UNICODE | capabilities.CAP_EXTENDED_SECURITY, true},
		{"extended security not negotiated", capabilities.CAP_UNICODE, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			treeResponse := commands.NewTreeConnectAndxResponse()
			tr := &capturingTransport{response: marshalResponse(t, treeResponse)}
			c := newSessionClient(tr)
			c.Connection.Server.Host = net.IPv4(192, 0, 2, 10)
			c.Connection.Server.Capabilities = test.caps
			if err := c.TreeConnect("IPC$"); err != nil {
				t.Fatalf("TreeConnect: %v", err)
			}
			if got := sentFlags2(t, tr.sent).IsExtendedSecurity(); got != test.want {
				t.Errorf("TREE_CONNECT extended-security flag = %v, want %v", got, test.want)
			}

			closeResponse := commands.NewCloseResponse()
			tr.response = marshalResponse(t, closeResponse)
			if err := c.CloseFile(client.FID(1)); err != nil {
				t.Fatalf("CloseFile: %v", err)
			}
			if got := sentFlags2(t, tr.sent).IsExtendedSecurity(); got != test.want {
				t.Errorf("file-I/O extended-security flag = %v, want %v", got, test.want)
			}
		})
	}
}
