package client

import (
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// TestNewFromTransport_NoNegotiate verifies the handoff constructor builds a
// client with initialized connection state and does not advance the MessageId
// (no NEGOTIATE has occurred yet).
func TestNewFromTransport_NoNegotiate(t *testing.T) {
	c := NewFromTransport(nil, net.ParseIP("10.0.0.1"), 445)
	if c.Connection == nil || c.Connection.Server == nil {
		t.Fatal("NewFromTransport did not initialize connection state")
	}
	if c.Connection.MessageId != 0 {
		t.Errorf("MessageId = %d, want 0 before negotiate", c.Connection.MessageId)
	}
	if c.GetPort() != 445 {
		t.Errorf("port = %d, want 445", c.GetPort())
	}
}

// TestApplyNegotiateResponse verifies that a handed-over NEGOTIATE response
// populates the connection/server state and advances the MessageId to 1, since
// NEGOTIATE occupies MessageId 0.
func TestApplyNegotiateResponse(t *testing.T) {
	c := NewFromTransport(nil, net.ParseIP("10.0.0.1"), 445)

	resp := commands.NewNegotiateResponse()
	resp.DialectRevision = dialects.SMB2_DIALECT_2_0_2
	resp.MaxReadSize = 0x10000
	resp.MaxWriteSize = 0x20000
	resp.SecurityBuffer = []byte{0x01, 0x02, 0x03}

	c.ApplyNegotiateResponse(resp)

	if c.Connection.Dialect != dialects.SMB2_DIALECT_2_0_2 {
		t.Errorf("Dialect = 0x%04x, want 0x%04x", uint16(c.Connection.Dialect), uint16(dialects.SMB2_DIALECT_2_0_2))
	}
	if c.Connection.Server.MaxReadSize != 0x10000 || c.Connection.Server.MaxWriteSize != 0x20000 {
		t.Errorf("max sizes not copied: read=%d write=%d", c.Connection.Server.MaxReadSize, c.Connection.Server.MaxWriteSize)
	}
	if len(c.Connection.Server.SecurityBuffer) != 3 {
		t.Errorf("security buffer not copied: %v", c.Connection.Server.SecurityBuffer)
	}
	if c.Connection.MessageId != 1 {
		t.Errorf("MessageId = %d, want 1 after applying negotiate", c.Connection.MessageId)
	}
}

// TestApplyNegotiateResponse_PreservesAdvancedMessageId verifies the MessageId is
// not reset when it has already advanced (the Negotiate path, where MessageId 0
// has been consumed and the counter is already at 1).
func TestApplyNegotiateResponse_PreservesAdvancedMessageId(t *testing.T) {
	c := NewFromTransport(nil, net.ParseIP("10.0.0.1"), 445)
	c.Connection.MessageId = 1 // as if NEGOTIATE was already sent on this client

	resp := commands.NewNegotiateResponse()
	resp.DialectRevision = dialects.SMB2_DIALECT_2_0_2
	c.ApplyNegotiateResponse(resp)

	if c.Connection.MessageId != 1 {
		t.Errorf("MessageId = %d, want it left at 1", c.Connection.MessageId)
	}
}
