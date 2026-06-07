package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// shareLevelSetupClient returns a client whose scripted transport replies to a
// share-level session setup with the given UID (share-level completes in one round
// trip, so a single canned response suffices).
func shareLevelSetupClient(uid uint16) (*client.Client, *scriptedTransport) {
	resp := message.NewMessage()
	resp.Header.SetFlags(flags.FLAGS_REPLY)
	resp.Header.UID = uid
	resp.AddCommand(commands.NewSessionSetupAndxResponse())
	raw, _ := resp.Marshal()

	tr := &scriptedTransport{response: raw}
	c := &client.Client{
		Transport:  tr,
		Connection: &client.Connection{Server: &client.Server{SecurityMode: 0}},
	}
	return c, tr
}

func TestSessionSetupRegistersSession(t *testing.T) {
	const uid = 0x0801
	c, _ := shareLevelSetupClient(uid)

	if err := c.SessionSetup(&credentials.Credentials{Username: "alice"}); err != nil {
		t.Fatalf("SessionSetup: %v", err)
	}
	if c.Connection.SessionTable[uid] != c.Session {
		t.Errorf("session not registered in SessionTable under UID 0x%04x", uid)
	}
}

func TestSessionSetupReusesSession(t *testing.T) {
	const uid = 0x0900
	c, tr := shareLevelSetupClient(uid)

	creds := &credentials.Credentials{Domain: "CORP", Username: "bob", Password: "pw"}
	if err := c.SessionSetup(creds); err != nil {
		t.Fatalf("first SessionSetup: %v", err)
	}
	first := c.Session
	sendsAfterFirst := tr.sendCount

	// A second setup with equal (but distinct-pointer) credentials must reuse the
	// registered session without another authentication exchange.
	if err := c.SessionSetup(&credentials.Credentials{Domain: "CORP", Username: "bob", Password: "pw"}); err != nil {
		t.Fatalf("second SessionSetup: %v", err)
	}
	if c.Session != first {
		t.Error("expected the second SessionSetup to reuse the existing session")
	}
	if tr.sendCount != sendsAfterFirst {
		t.Errorf("expected no additional messages on reuse, sent %d then %d", sendsAfterFirst, tr.sendCount)
	}
}

func TestSessionSetupDifferentCredentialsDoesNotReuse(t *testing.T) {
	const uid = 0x0A00
	c, tr := shareLevelSetupClient(uid)

	if err := c.SessionSetup(&credentials.Credentials{Username: "alice"}); err != nil {
		t.Fatalf("first SessionSetup: %v", err)
	}
	sendsAfterFirst := tr.sendCount

	// Different credentials must trigger a fresh exchange (more sends).
	if err := c.SessionSetup(&credentials.Credentials{Username: "carol"}); err != nil {
		t.Fatalf("second SessionSetup: %v", err)
	}
	if tr.sendCount == sendsAfterFirst {
		t.Error("expected a fresh authentication exchange for different credentials")
	}
}

func TestLogoffRemovesSessionFromTable(t *testing.T) {
	tr := &capturingTransport{response: marshalResponse(t, commands.NewLogoffAndxResponse())}
	c := newSessionClient(tr)
	uid := c.Session.SessionUID
	c.Connection.SessionTable = map[uint16]*client.Session{uid: c.Session}

	if err := c.Logoff(); err != nil {
		t.Fatalf("Logoff: %v", err)
	}
	if _, ok := c.Connection.SessionTable[uid]; ok {
		t.Errorf("session UID 0x%04x still present in SessionTable after Logoff", uid)
	}
	if c.Session != nil {
		t.Error("expected Client.Session to be nil after Logoff")
	}
}
