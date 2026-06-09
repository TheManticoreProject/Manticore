package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
)

// cannedResponse wraps a command body in a server response message (status code
// and header TreeId/SessionId controllable) and returns its wire bytes.
func cannedResponse(t *testing.T, cmd command_interface.CommandInterface, status uint32, treeId uint32, sessionId uint64) []byte {
	t.Helper()
	m := message.NewMessage()
	m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.Header.Status = status
	m.Header.TreeId = treeId
	m.Header.SessionId = sessionId
	m.SetCommand(cmd)
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("building canned response: %v", err)
	}
	return wire
}

func TestTreeConnectAndLifecycle(t *testing.T) {
	tcResp := commands.NewTreeConnectResponse()
	tcResp.ShareType = commands.SMB2_SHARE_TYPE_DISK

	ft := &fakeTransport{responses: [][]byte{
		cannedResponse(t, tcResp, 0, 0x5, 0),                              // TREE_CONNECT -> TreeId 5
		cannedResponse(t, commands.NewTreeDisconnectResponse(), 0, 0x5, 0), // TREE_DISCONNECT
		cannedResponse(t, commands.NewLogoffResponse(), 0, 0, 0x99),        // LOGOFF
	}}
	c := newTestClient(ft)
	// A session is required; install one directly (NTLM setup needs a live server).
	c.Session = &Session{Client: c, SessionId: 0x99}
	c.Connection.SessionTable[0x99] = c.Session

	if err := c.TreeConnect("share"); err != nil {
		t.Fatalf("TreeConnect: %v", err)
	}
	if c.Session.TreeId != 0x5 {
		t.Errorf("TreeId = %d, want 5", c.Session.TreeId)
	}
	if _, ok := c.Connection.TreeConnectTable[0x5]; !ok {
		t.Errorf("tree connect not registered in table")
	}

	if err := c.TreeDisconnect(); err != nil {
		t.Fatalf("TreeDisconnect: %v", err)
	}
	if c.Session.TreeId != 0 {
		t.Errorf("TreeId after disconnect = %d, want 0", c.Session.TreeId)
	}
	if _, ok := c.Connection.TreeConnectTable[0x5]; ok {
		t.Errorf("tree connect still in table after disconnect")
	}

	if err := c.Logoff(); err != nil {
		t.Fatalf("Logoff: %v", err)
	}
	if c.Session != nil {
		t.Errorf("session not cleared after logoff")
	}
}

func TestTreeConnectRequiresSession(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	if err := c.TreeConnect("share"); err == nil {
		t.Errorf("expected TreeConnect without a session to error")
	}
}
