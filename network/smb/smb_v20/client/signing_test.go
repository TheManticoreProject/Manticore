package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestSignAndVerifyRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef") // 16-byte session key

	m := message.NewMessage()
	m.Header.MessageId = 3
	m.SetCommand(commands.NewTreeDisconnectRequest())
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	signMessage(key, wire)

	// The SIGNED flag must be set after signing.
	decoded := message.NewMessage()
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.Header.Flags.IsSigned() {
		t.Errorf("SMB2_FLAGS_SIGNED not set after signing")
	}

	// The signature must verify with the same key and fail with a different key.
	if !verifySignature(key, wire) {
		t.Errorf("verifySignature failed for a correctly signed message")
	}
	if verifySignature([]byte("WRONGKEYWRONGKEY"), wire) {
		t.Errorf("verifySignature succeeded with the wrong key")
	}

	// Tampering with the body must invalidate the signature.
	wire[len(wire)-1] ^= 0xFF
	if verifySignature(key, wire) {
		t.Errorf("verifySignature succeeded for a tampered message")
	}
}

func TestVerifyRejectsShortMessage(t *testing.T) {
	if verifySignature([]byte("key"), make([]byte, 10)) {
		t.Errorf("verifySignature should reject a sub-header-length message")
	}
}

// TestSendReceiveEnforcesSigning verifies that when session signing is active a
// final response is accepted only when correctly signed: an unsigned response is
// rejected (MS-SMB2 3.2.5.1.3), and a properly signed one is accepted.
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
		signMessage(key, raw)
		ft := &fakeTransport{responses: [][]byte{raw}}
		c := newTestClient(ft)
		signingSession(c)

		if err := c.CloseFile(types.SMB2_FILEID{}); err != nil {
			t.Fatalf("expected a correctly signed response to be accepted, got: %v", err)
		}
	})
}
