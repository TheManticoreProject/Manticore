package client

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	smb2 "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// Client is the generic, version-agnostic SMB client. It negotiates a dialect
// once and routes every operation to a version-specific backend (SMB1/SMB2/…).
// The surface is flat: the client holds the current session and tree, mirroring
// the underlying engines.
type Client struct {
	backend Backend
	host    string
	port    int
	opts    Options
}

// Dial connects to host:port, negotiates a dialect according to opts, and
// returns a ready-to-authenticate Client.
//
// Selection is driven by opts.Preferred (highest priority first; defaults to all
// supported versions, best first) and opts.Policy:
//
//   - PolicyStrictOrder (default) tries the preferred versions in order and
//     returns the first the server accepts, reconnecting between attempts. It
//     can select a lower dialect even when the server supports a higher one.
//   - PolicyHighestInSet performs a single SMB1 multi-protocol negotiate and
//     uses the server's highest-supported dialect within the set.
func Dial(host string, port int, opts Options) (*Client, error) {
	ip, err := resolveHost(host)
	if err != nil {
		return nil, err
	}

	prefs := opts.Preferred
	if len(prefs) == 0 {
		prefs = defaultPreference()
	}

	if opts.Policy == PolicyHighestInSet {
		return dialHighestInSet(host, ip, port, opts, prefs)
	}
	return dialStrictOrder(host, ip, port, opts, prefs)
}

// dialSMB2 connects with the SMB2 engine and performs its native negotiate.
func dialSMB2(ip net.IP, host string, port int, opts Options) (*Client, error) {
	engine := smb2.NewClientUsingTCPTransport(ip, port)
	if opts.Workstation != "" {
		engine.Workstation = opts.Workstation
	}
	if err := engine.Connect(ip, port); err != nil {
		return nil, fmt.Errorf("SMB2 connect/negotiate failed: %w", err)
	}
	return &Client{backend: newSMB2Backend(engine), host: host, port: port, opts: opts}, nil
}

// dialSMB1 connects with the SMB1 engine and performs its native negotiate.
func dialSMB1(ip net.IP, host string, port int, opts Options) (*Client, error) {
	engine := smb1.NewClientUsingTCPTransport(ip, port)
	// The SMB1 session setup requires the NativeOS / NativeLanMan fields to be set.
	engine.NativeOS = "Manticore"
	engine.NativeLanMan = "Manticore"
	if opts.Workstation != "" {
		engine.Workstation = opts.Workstation
	}
	if err := engine.Connect(ip, port); err != nil {
		return nil, fmt.Errorf("SMB1 connect/negotiate failed: %w", err)
	}
	return &Client{backend: newSMB1Backend(engine), host: host, port: port, opts: opts}, nil
}

// Dialect reports the negotiated protocol version.
func (c *Client) Dialect() smb.SMBProtocolVersion { return c.backend.Dialect() }

// Login authenticates a session with the server.
func (c *Client) Login(creds *credentials.Credentials) error { return c.backend.Login(creds) }

// TreeConnect connects to a share by name (e.g. "C$"), making it the current tree.
func (c *Client) TreeConnect(share string) error { return c.backend.TreeConnect(share) }

// OpenFile opens or creates a file on the current tree and returns a handle.
func (c *Client) OpenFile(path string, opts OpenOptions) (FileHandle, error) {
	return c.backend.OpenFile(path, opts)
}

// ReadFile reads up to n bytes from the open file starting at off.
func (c *Client) ReadFile(h FileHandle, off uint64, n uint32) ([]byte, error) {
	return c.backend.ReadFile(h, off, n)
}

// WriteFile writes data to the open file at off and returns the bytes written.
func (c *Client) WriteFile(h FileHandle, off uint64, data []byte) (uint32, error) {
	return c.backend.WriteFile(h, off, data)
}

// CloseFile closes a handle returned by OpenFile.
func (c *Client) CloseFile(h FileHandle) error { return c.backend.CloseFile(h) }

// ListDirectory enumerates entries of a directory on the current tree.
func (c *Client) ListDirectory(path, pattern string) ([]FileInfo, error) {
	return c.backend.ListDirectory(path, pattern)
}

// TreeDisconnect disconnects the current tree.
func (c *Client) TreeDisconnect() error { return c.backend.TreeDisconnect() }

// Logoff tears down the authenticated session.
func (c *Client) Logoff() error { return c.backend.Logoff() }

// Disconnect closes the underlying transport connection.
func (c *Client) Disconnect() error { return c.backend.Disconnect() }

// resolveHost accepts a literal IP or a hostname and returns an IP address.
func resolveHost(host string) (net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return ip, nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host %q: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses found for host %q", host)
	}
	return ips[0], nil
}
