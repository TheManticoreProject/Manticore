package client

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/transport"
)

// Client represents an SMB v1.0 client
type Client struct {
	Transport   transport.Transport
	TreeConnect *TreeConnect
	Session     *Session    // An established session between the client and server
	Connection  *Connection // An established SMB connection between the client and the server
}

// NewClientUsingNBTTransport creates a new SMB v1.0 client using NBT transport
//
// Returns:
//   - A pointer to the initialized SMB client
//   - An error if the client initialization fails
func NewClientUsingNBTTransport(host net.IP, port int) *Client {
	return &Client{
		Transport: transport.NewTransport("nbt"),
		Connection: &Connection{
			Host: host,
			Port: port,
		},
		TreeConnect: nil,
		Session:     nil,
	}
}

// Connect establishes a connection to an SMB server
//
// Returns:
//   - An error if the connection fails
func (c *Client) Connect(ipaddr net.IP, port int) error {
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

// SetHost sets the host IP address for the SMB client
func (c *Client) SetHost(host net.IP) {
	c.Connection.Host = host
}

// GetHost returns the current host IP address of the SMB client
func (c *Client) GetHost() net.IP {
	return c.Connection.Host
}

// SetPort sets the port number for the SMB client
func (c *Client) SetPort(port int) {
	c.Connection.Port = port
}

// GetPort returns the current port number of the SMB client
func (c *Client) GetPort() int {
	return c.Connection.Port
}
