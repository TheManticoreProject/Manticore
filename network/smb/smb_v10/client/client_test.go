package client_test

import (
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/securitymode"
)

// connectedMockTransport is a transport.Transport stub that reports itself as
// connected. Send/Receive are not expected to be reached by these tests.
type connectedMockTransport struct{}

func (m *connectedMockTransport) Connect(ipaddr net.IP, port int) error { return nil }
func (m *connectedMockTransport) Close() error                          { return nil }
func (m *connectedMockTransport) Send(data []byte) (int, error)         { return len(data), nil }
func (m *connectedMockTransport) Receive() ([]byte, error)              { return nil, nil }
func (m *connectedMockTransport) IsConnected() bool                     { return true }

// TestTreeConnectWithoutSession verifies that calling TreeConnect before a
// session has been established returns an error rather than panicking with a
// nil-pointer dereference on c.Session.
func TestTreeConnectWithoutSession(t *testing.T) {
	c := &client.Client{}

	err := c.TreeConnect("share")
	if err == nil {
		t.Fatal("expected an error when TreeConnect is called without a session, got nil")
	}
}

// TestSessionSetupWithNilCredentials verifies that SessionSetup returns an error
// rather than panicking with a nil-pointer dereference when no credentials are
// supplied.
func TestSessionSetupWithNilCredentials(t *testing.T) {
	// Configure the server security mode so SessionSetup takes the user-level
	// challenge/response path, which dereferences s.Credentials directly.
	c := &client.Client{
		Transport: &connectedMockTransport{},
		Connection: &client.Connection{
			Server: &client.Server{
				SecurityMode: securitymode.NEGOTIATE_USER_SECURITY | securitymode.NEGOTIATE_ENCRYPT_PASSWORDS,
			},
		},
	}

	err := c.SessionSetup(nil)
	if err == nil {
		t.Fatal("expected an error when SessionSetup is called with nil credentials, got nil")
	}
}
