package client

import (
	"net"
	"testing"
)

// TestNewFromTransport_NoNegotiate verifies the handoff constructor builds an
// SMB1 client with initialized connection state over a (here nil) transport,
// without connecting or negotiating. ApplyNegotiateResponse, the other half of
// the handoff, is exercised through the Negotiate path in the client suite.
func TestNewFromTransport_NoNegotiate(t *testing.T) {
	c := NewFromTransport(nil, net.ParseIP("10.0.0.1"), 445)
	if c.Connection == nil || c.Connection.Server == nil {
		t.Fatal("NewFromTransport did not initialize connection state")
	}
	if c.Session != nil {
		t.Error("NewFromTransport should not create a session")
	}
	if c.GetHost().String() != "10.0.0.1" || c.GetPort() != 445 {
		t.Errorf("host/port = %s/%d, want 10.0.0.1/445", c.GetHost(), c.GetPort())
	}
}
