package client

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// NewClientUsingNBTTransport creates a new SMB v1.0 client using NBT transport
//
// Returns:
//   - A pointer to the initialized SMB client
//   - An error if the client initialization fails
func NewClientUsingNBTTransport(host net.IP, port int) *Client {
	return &Client{
		Transport: transport.NewTransport("nbt"),
		Connection: &Connection{
			Server: &Server{
				Host: host,
				Port: port,
			},
		},
		Session: nil,
	}
}

// NewClientUsingTCPTransport creates a new SMB v1.0 client using TCP transport
//
// Returns:
//   - A pointer to the initialized SMB client
//   - An error if the client initialization fails
func NewClientUsingTCPTransport(host net.IP, port int) *Client {
	return &Client{
		Transport: transport.NewTransport("tcp"),

		Connection: &Connection{
			Server: &Server{
				Host: host,
				Port: port,
			},
		},
		Session: nil,
	}
}

// Connect establishes a connection to an SMB server
//
// Returns:
func (c *Client) Connect(ipaddr net.IP, port int) error {
	// Record the connection address as the single source of truth so that later
	// operations (e.g. TreeConnect, which builds the UNC path from
	// Connection.Server.Host) target the same server this connection is made to,
	// rather than a stale value left by the constructor.
	c.Connection.Server.Host = ipaddr
	c.Connection.Server.Port = port

	err := c.Transport.Connect(ipaddr, port)
	if err != nil {
		return fmt.Errorf("failed to connect to SMB server: %v", err)
	}

	err = c.Negotiate()
	if err != nil {
		return fmt.Errorf("failed to negotiate with SMB server: %v", err)
	}

	return nil
}

// SessionSetup sets up a session with the SMB server
//
// Returns:
//   - An error if the session setup fails
func (c *Client) SessionSetup(creds *credentials.Credentials) error {
	// Reuse an already-authenticated session for the same credentials if one is
	// registered on this connection, avoiding a redundant authentication exchange.
	if existing := c.findSessionByCredentials(creds); existing != nil {
		c.Session = existing
		return nil
	}

	if c.Session == nil {
		c.Session = &Session{
			Client:      c,
			Credentials: creds,
		}
	}
	c.Session.Client = c
	if c.Session.Credentials == nil {
		c.Session.Credentials = creds
	}

	if err := c.Session.SessionSetup(); err != nil {
		return err
	}

	// Register the established session in the connection's session table, keyed by
	// the server-assigned UID, so it can be reused and is removed on Logoff.
	if c.Connection != nil {
		if c.Connection.SessionTable == nil {
			c.Connection.SessionTable = make(map[uint16]*Session)
		}
		c.Connection.SessionTable[c.Session.SessionUID] = c.Session
	}

	return nil
}

// findSessionByCredentials returns a session already registered on the connection
// whose credentials match creds, or nil if none is found. A nil creds (anonymous)
// is never matched, so anonymous setups always perform a fresh exchange.
func (c *Client) findSessionByCredentials(creds *credentials.Credentials) *Session {
	if creds == nil || c.Connection == nil {
		return nil
	}
	for _, session := range c.Connection.SessionTable {
		if session != nil && sameCredentials(session.Credentials, creds) {
			return session
		}
	}
	return nil
}

// sameCredentials reports whether two credential sets identify the same principal.
func sameCredentials(a, b *credentials.Credentials) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Domain == b.Domain &&
		a.Username == b.Username &&
		a.Password == b.Password &&
		a.LMHash == b.LMHash &&
		a.NTHash == b.NTHash
}

// SetHost sets the host IP address for the SMB client
func (c *Client) SetHost(host net.IP) {
	c.Connection.Server.Host = host
}

// GetHost returns the current host IP address of the SMB client
func (c *Client) GetHost() net.IP {
	return c.Connection.Server.Host
}

// SetPort sets the port number for the SMB client
func (c *Client) SetPort(port int) {
	c.Connection.Server.Port = port
}

// GetPort returns the current port number of the SMB client
func (c *Client) GetPort() int {
	return c.Connection.Server.Port
}
