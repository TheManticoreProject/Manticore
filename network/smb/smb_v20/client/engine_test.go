package client

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// cannedResponseWithMessageId is cannedResponse with an explicit MessageId, used
// to exercise the engine's request/response matching. The fakeTransport leaves a
// non-zero MessageId untouched.
func cannedResponseWithMessageId(t *testing.T, cmd command_interface.CommandInterface, status uint32, treeId uint32, sessionId uint64, messageId uint64) []byte {
	t.Helper()
	m := message.NewMessage()
	m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.Header.Status = status
	m.Header.TreeId = treeId
	m.Header.SessionId = sessionId
	m.Header.MessageId = types.UINT64(messageId)
	m.SetCommand(cmd)
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("building canned response: %v", err)
	}
	return wire
}

// TestSendReceiveSkipsUnsolicitedOplockBreak verifies that an OPLOCK_BREAK
// notification (reserved MessageId 0xFFFFFFFFFFFFFFFF) interleaved before the
// real response is skipped rather than consumed as the operation's response.
func TestSendReceiveSkipsUnsolicitedOplockBreak(t *testing.T) {
	oplockBreak := commands.NewOplockBreakResponse()
	ft := &fakeTransport{responses: [][]byte{
		cannedResponseWithMessageId(t, oplockBreak, 0, 0x5, 0x99, unsolicitedMessageId),
		cannedResponse(t, commands.NewCloseResponse(), 0, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)

	if err := c.CloseFile(types.SMB2_FILEID{}); err != nil {
		t.Fatalf("CloseFile should skip the interleaved oplock break and use the real response, got: %v", err)
	}
}

// TestSendReceiveRejectsMismatchedMessageId verifies that a response whose
// MessageId does not match the request is rejected as a desynchronized stream.
func TestSendReceiveRejectsMismatchedMessageId(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{
		cannedResponseWithMessageId(t, commands.NewCloseResponse(), 0, 0x5, 0x99, 999),
	}}
	c := withConnectedTree(ft)

	if err := c.CloseFile(types.SMB2_FILEID{}); err == nil {
		t.Fatal("CloseFile should reject a response with a mismatched MessageId")
	}
}

// TestSendReceiveRejectsMismatchedCommand verifies that a response whose command
// code does not match the request is rejected.
func TestSendReceiveRejectsMismatchedCommand(t *testing.T) {
	// Reply to a CLOSE request with a READ response (matching MessageId 0).
	ft := &fakeTransport{responses: [][]byte{
		cannedResponse(t, commands.NewReadResponse(), 0, 0x5, 0x99),
	}}
	c := withConnectedTree(ft)

	if err := c.CloseFile(types.SMB2_FILEID{}); err == nil {
		t.Fatal("CloseFile should reject a response carrying a different command code")
	}
}

// TestSendReceiveRejectsUnencryptedWhenEncryptDataSet verifies that a plaintext
// response is rejected when Session.EncryptData is true (MS-SMB2 3.2.5.1.1).
func TestSendReceiveRejectsUnencryptedWhenEncryptDataSet(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}
	ft := &fakeTransport{responses: [][]byte{
		cannedResponse(t, commands.NewCloseResponse(), 0, 0x5, 0x99),
	}}
	c := newTestClient(ft)
	c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1
	c.Connection.Cipher = commands.SMB2_ENCRYPTION_AES128_GCM
	c.Session = &Session{
		Client:        c,
		SessionId:     0x99,
		TreeId:        0x5,
		EncryptionKey: key,
		DecryptionKey: key,
		EncryptData:   true,
	}
	c.Connection.SessionTable[0x99] = c.Session

	err := c.CloseFile(types.SMB2_FILEID{})
	if err == nil {
		t.Fatal("CloseFile should reject an unencrypted response when Session.EncryptData is true")
	}
	if !strings.Contains(err.Error(), "session requires encryption") {
		t.Fatalf("unexpected error: %v", err)
	}
}
