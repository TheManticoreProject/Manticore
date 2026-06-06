package client

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/transport"
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
func (c *Client) SessionSetup(credentials *credentials.Credentials) error {
	if c.Session == nil {
		c.Session = &Session{
			Client:      c,
			Credentials: credentials,
		}
	}
	c.Session.Client = c

	return c.Session.SessionSetup()
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
