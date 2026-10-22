package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/security"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestSendReceiveEnforcesSigning(t *testing.T) {
	key := []byte("0123456789abcdef")

	signingSession := func(c *Client) {
		c.Session = &Session{Client: c, SessionId: 0x99, TreeId: 0x5, SigningActive: true, SigningKey: key}
		c.Connection.SessionTable[0x99] = c.Session
	}

	t.Run("unsigned response rejected", func(t *testing.T) {
		ft := &fakeTransport{responses: [][]byte{
			cannedResponse(t, commands.NewCloseResponse(), 0, 0x5, 0x99), // unsigned
		}}
		c := newTestClient(ft)
		signingSession(c)

		if err := c.CloseFile(types.SMB2_FILEID{}); err == nil {
			t.Fatal("expected an unsigned response to be rejected when signing is active")
		}
	})

	t.Run("signed response accepted", func(t *testing.T) {
		raw := cannedResponse(t, commands.NewCloseResponse(), 0, 0x5, 0x99)
		security.Sign(dialects.SMB2_DIALECT_2_0_2, -1, key, raw)
		ft := &fakeTransport{responses: [][]byte{raw}}
		c := newTestClient(ft)
		signingSession(c)

		if err := c.CloseFile(types.SMB2_FILEID{}); err != nil {
			t.Fatalf("expected a correctly signed response to be accepted, got: %v", err)
		}
	})
}
