package client

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
)

// newClient builds a client over the given transport, targeting host:port.
func newClient(t transport.Transport, host net.IP, port int) *Client {
	return &Client{
		Transport: t,
		Connection: &Connection{
			Server:           &Server{Host: host, Port: port},
			SessionTable:     make(map[uint64]*Session),
			TreeConnectTable: make(map[uint32]*TreeConnect),
		},
	}
}

// NewClientUsingTCPTransport creates a new SMB 2.0 client using direct TCP transport.
func NewClientUsingTCPTransport(host net.IP, port int) *Client {
	return newClient(transport.NewTransport("tcp"), host, port)
}

// NewClientUsingNBTTransport creates a new SMB 2.0 client using NetBIOS transport.
func NewClientUsingNBTTransport(host net.IP, port int) *Client {
	return newClient(transport.NewTransport("nbt"), host, port)
}

// NewFromTransport builds an SMB 2.0 client over an already-connected transport,
// without connecting or negotiating. It is the handoff entry point used by the
// generic SMB client (network/smb/client), which performs a multi-protocol
// negotiate and then hands the live transport to this engine. The caller drives
// negotiation either via Negotiate (a native SMB2 NEGOTIATE over t) or by
// applying an already-exchanged response with ApplyNegotiateResponse.
func NewFromTransport(t transport.Transport, host net.IP, port int) *Client {
	return newClient(t, host, port)
}

// Connect establishes the transport connection and performs SMB2 negotiation.
func (c *Client) Connect(ipaddr net.IP, port int) error {
	c.Connection.Server.Host = ipaddr
	c.Connection.Server.Port = port

	if err := c.Transport.Connect(ipaddr, port); err != nil {
		return fmt.Errorf("failed to connect to SMB2 server: %w", err)
	}

	if err := c.Negotiate(); err != nil {
		return fmt.Errorf("failed to negotiate with SMB2 server: %w", err)
	}

	return nil
}

// Disconnect closes the underlying transport connection. It does not send any
// SMB2 commands; call TreeDisconnect and Logoff first for a clean teardown.
func (c *Client) Disconnect() error {
	if c.Transport == nil {
		return nil
	}
	return c.Transport.Close()
}

// SetHost sets the target server IP address.
func (c *Client) SetHost(host net.IP) { c.Connection.Server.Host = host }

// GetHost returns the target server IP address.
func (c *Client) GetHost() net.IP { return c.Connection.Server.Host }

// SetPort sets the target server port.
func (c *Client) SetPort(port int) { c.Connection.Server.Port = port }

// GetPort returns the target server port.
func (c *Client) GetPort() int { return c.Connection.Server.Port }

// EnableEncryption turns on SMB 3.x per-message encryption for the current
// session, wrapping every subsequent request in an SMB2 TRANSFORM_HEADER. It
// requires an SMB 3.x session whose encryption keys have been derived; it
// returns an error otherwise. The server transparently encrypts its replies once
// it receives an encrypted request.
func (c *Client) EnableEncryption() error {
	if c.Session == nil {
		return fmt.Errorf("cannot enable encryption: no session established")
	}
	if !isSMB3Dialect(c.Connection.Dialect) {
		return fmt.Errorf("cannot enable encryption: dialect %s does not support SMB3 encryption", c.Connection.Dialect)
	}
	if len(c.Session.EncryptionKey) == 0 {
		return fmt.Errorf("cannot enable encryption: no encryption key derived")
	}
	c.Session.EncryptData = true
	return nil
}

// IsEncryptionActive reports whether the current session encrypts its traffic.
func (c *Client) IsEncryptionActive() bool {
	return c.Session != nil && c.Session.EncryptData
}
