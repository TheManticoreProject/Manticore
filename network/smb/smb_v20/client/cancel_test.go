package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestCancelRequiresPendingAsync(t *testing.T) {
	c := withConnectedTree(&fakeTransport{})
	if err := c.Cancel(); err == nil {
		t.Fatal("Cancel should fail when no async operation is pending")
	}
}

func TestCancelSetsAsyncFields(t *testing.T) {
	ft := &fakeTransport{}
	c := withConnectedTree(ft)
	c.setPendingAsync(42, 0xBEEF)

	if err := c.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if len(ft.sent) != 1 {
		t.Fatalf("expected 1 sent frame, got %d", len(ft.sent))
	}

	sent := ft.sent[0]
	if len(sent) < 64 {
		t.Fatalf("sent frame is %d bytes, want at least 64", len(sent))
	}
	msgFlags := flags.Flags(uint32(sent[16]) | uint32(sent[17])<<8 | uint32(sent[18])<<16 | uint32(sent[19])<<24)
	if !msgFlags.IsAsync() {
		t.Error("CANCEL must set SMB2_FLAGS_ASYNC_COMMAND")
	}
	gotMsgId := types.UINT64(uint64(sent[24]) | uint64(sent[25])<<8 | uint64(sent[26])<<16 | uint64(sent[27])<<24 |
		uint64(sent[28])<<32 | uint64(sent[29])<<40 | uint64(sent[30])<<48 | uint64(sent[31])<<56)
	if gotMsgId != 42 {
		t.Errorf("MessageId = %d, want 42", gotMsgId)
	}
}

func TestCancelEncryptsWhenSessionEncrypts(t *testing.T) {
	key := mustHex(t, "629BCBC54422A0F572B97F45989B6073")
	ft := &fakeTransport{}
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
	c.setPendingAsync(7, 0xABCD)

	if err := c.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if len(ft.sent) != 1 {
		t.Fatalf("expected 1 sent frame, got %d", len(ft.sent))
	}

	sent := ft.sent[0]
	if !isTransformHeader(sent) {
		t.Fatal("CANCEL must be encrypted (TRANSFORM_HEADER) when EncryptData is set")
	}
}

func TestCancelSignsWhenSigningActiveNoEncryption(t *testing.T) {
	key := []byte("0123456789abcdef")
	ft := &fakeTransport{}
	c := newTestClient(ft)
	c.Connection.Dialect = dialects.SMB2_DIALECT_2_0_2
	c.Session = &Session{
		Client:       c,
		SessionId:    0x99,
		TreeId:       0x5,
		SigningKey:    key,
		SigningActive: true,
	}
	c.Connection.SessionTable[0x99] = c.Session
	c.setPendingAsync(7, 0xABCD)

	if err := c.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if len(ft.sent) != 1 {
		t.Fatalf("expected 1 sent frame, got %d", len(ft.sent))
	}

	sent := ft.sent[0]
	if isTransformHeader(sent) {
		t.Fatal("CANCEL should be signed, not encrypted, when EncryptData is false")
	}
	if !verifySignature(key, sent) {
		t.Error("CANCEL must be signed when SigningActive is set")
	}
}
