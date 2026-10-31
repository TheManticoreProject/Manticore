package client

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// TestSendReceiveCompoundRejectsPlaintextAfterEncryptedBreak verifies that an
// encrypted server-initiated frame (a break) read ahead of a compounded response
// does not make the response count as encrypted: on a session that requires
// encryption, a plaintext, unsigned compound response must still be refused.
func TestSendReceiveCompoundRejectsPlaintextAfterEncryptedBreak(t *testing.T) {
	key := mustHex(t, "629BCBC54422A0F572B97F45989B6073")
	encryptingClient := func(ft *fakeTransport) *Client {
		c := newTestClient(ft)
		c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1
		c.Connection.Cipher = commands.SMB2_ENCRYPTION_AES128_GCM
		c.Session = &Session{
			Client: c, SessionId: 0x99, TreeId: 0x5,
			SigningActive: true, SigningKey: key,
			EncryptionKey: key, DecryptionKey: key, EncryptData: true,
		}
		c.Connection.SessionTable[0x99] = c.Session
		return c
	}

	encryptedBreak, err := encryptingClient(&fakeTransport{}).encryptMessage(buildOplockBreakWire(t))
	if err != nil {
		t.Fatalf("encryptMessage: %v", err)
	}
	ft := &fakeTransport{responses: [][]byte{
		encryptedBreak,
		createQueryCloseResponse(t, []byte{0x01}), // plaintext and unsigned
	}}
	c := encryptingClient(ft)

	_, err = c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0)
	if err == nil || !strings.Contains(err.Error(), "not encrypted") {
		t.Fatalf("CreateQueryInfoClose = %v, want the plaintext response refused as not encrypted", err)
	}
}

// TestSendReceiveCompoundVerifiesSignatureAfterEncryptedBreak verifies the same
// on a session that signs but does not require encryption: an encrypted break
// read first must not exempt the plaintext compound response from signature
// verification.
func TestSendReceiveCompoundVerifiesSignatureAfterEncryptedBreak(t *testing.T) {
	key := mustHex(t, "629BCBC54422A0F572B97F45989B6073")
	signingClient := func(ft *fakeTransport) *Client {
		c := newTestClient(ft)
		c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1
		c.Connection.Cipher = commands.SMB2_ENCRYPTION_AES128_GCM
		c.Session = &Session{
			Client: c, SessionId: 0x99, TreeId: 0x5,
			SigningActive: true, SigningKey: key,
			EncryptionKey: key, DecryptionKey: key,
		}
		c.Connection.SessionTable[0x99] = c.Session
		return c
	}

	encryptedBreak, err := signingClient(&fakeTransport{}).encryptMessage(buildOplockBreakWire(t))
	if err != nil {
		t.Fatalf("encryptMessage: %v", err)
	}
	ft := &fakeTransport{responses: [][]byte{
		encryptedBreak,
		createQueryCloseResponse(t, []byte{0x01}), // plaintext and unsigned
	}}
	c := signingClient(ft)

	_, err = c.CreateQueryInfoClose("x", 0x00100081, 0x07, 0x00000001, 0, commands.SMB2_0_INFO_FILE, 0x12, 0)
	if err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("CreateQueryInfoClose = %v, want the unsigned response refused", err)
	}
}
