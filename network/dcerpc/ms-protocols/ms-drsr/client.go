// Package msdrsr implements high-level MS-DRSR (Directory Replication Service Remote
// Protocol) workflows over the drsuapi DCE/RPC interface
// (e3514235-4b06-11d1-ab04-00c04fc2dcd2 v4.0), carried over ncacn_ip_tcp.
//
// Unlike the named-pipe protocols (ms-srvs, ms-rrp), drsuapi has no fixed endpoint: the
// server listens on a dynamic TCP port that the client resolves through the endpoint
// mapper on TCP/135. A Client therefore owns its own transport lifecycle — it resolves
// the endpoint, dials it, authenticates with NTLM at packet-privacy level (drsuapi
// requires sign+seal), binds the drsuapi abstract syntax, and performs the IDL_DRSBind
// capability handshake — rather than borrowing an established SMB session.
//
// This file covers the connection lifecycle (Connect/Close, i.e. IDL_DRSBind /
// IDL_DRSUnbind and the DRS_EXTENSIONS negotiation). The replication workflows
// (DRSCrackNames, DRSGetNCChanges, DCSync) build on the bound handle in later files.
//
// References:
//   - [MS-DRSR] 4.1.3 IDL_DRSBind, 4.1.25 IDL_DRSUnbind, 5.39 DRS_EXTENSIONS_INT
//   - [MS-RPCE] 2.1.1.1 ncacn_ip_tcp; endpoint mapper ept_map ([C706] Appendix O)
package msdrsr

import (
	"fmt"
	"time"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/msproto"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// DefaultTimeout bounds the endpoint-mapper and drsuapi TCP dials and reads when the
// caller does not set one with SetTimeout.
const DefaultTimeout = 10 * time.Second

// Client is an MS-DRSR client over ncacn_ip_tcp. The zero value is not usable; build one
// with New. It is not safe for concurrent use.
type Client struct {
	host    string
	port    int // explicit drsuapi port; 0 means resolve via the endpoint mapper
	creds   *credentials.Credentials
	timeout time.Duration

	rpc       *dcerpcclient.Client
	handle    structures.DRS_HANDLE
	serverExt *structures.DRS_EXTENSIONS_INT
	sourceDSA structures.UUID // source DSA GUID for GetNCChanges; zero (NULL) by default
	bound     bool

	// onReplicationProgress, when non-nil, is invoked once per replicated page during
	// full-NC replication (ReplicateNC) with the cumulative number of objects received
	// so far. Registered via SetReplicationProgress.
	onReplicationProgress func(objects int)
}

// compile-time assertion that Client satisfies the session contract.
var _ msproto.Session = (*Client)(nil)

// New returns an MS-DRSR client for the given host (IP or hostname) and credentials. By
// default the drsuapi TCP endpoint is resolved through the endpoint mapper on TCP/135;
// call SetPort to skip resolution and dial a known port directly.
func New(host string, creds *credentials.Credentials) *Client {
	return &Client{host: host, creds: creds, timeout: DefaultTimeout}
}

// Interface reports the DCE/RPC abstract syntax MS-DRSR speaks (drsuapi v4.0).
func (c *Client) Interface() syntax.SyntaxID { return drsuapi.SyntaxID() }

// IsConnected reports whether Connect has succeeded and Close has not yet run.
func (c *Client) IsConnected() bool { return c.bound }

// SetTimeout bounds each TCP dial and read. It must be called before Connect.
func (c *Client) SetTimeout(d time.Duration) { c.timeout = d }

// SetPort overrides endpoint-mapper resolution and dials drsuapi on the given TCP port
// directly. It must be called before Connect.
func (c *Client) SetPort(port int) { c.port = port }

// Connect resolves the drsuapi endpoint (unless a port was set), dials it over
// ncacn_ip_tcp, authenticates with NTLM at packet-privacy level, binds the drsuapi
// abstract syntax, and performs the IDL_DRSBind capability handshake. On success the
// negotiated context handle and the server's extensions are available via Handle and
// ServerExtensions, and the connection is ready for replication calls.
//
// The endpoint resolution, NTLM packet-privacy setup (drsuapi rejects lower levels), and
// bind are handled by an msproto.TCPBinder; only the drsuapi-specific IDL_DRSBind
// capability handshake lives here.
func (c *Client) Connect() error {
	if c.bound {
		return fmt.Errorf("msdrsr: already connected")
	}

	binder := msproto.NewTCPBinder(c.host, c.port, c.creds, c.timeout)
	rpc, _, err := binder.Bind(drsuapi.SyntaxID())
	if err != nil {
		return fmt.Errorf("msdrsr: %w", err)
	}

	clientGUID := structures.NTDSAPIClientGUID()
	clientExt := structures.DefaultClientExtensions().ToExtensions()
	serverExt, handle, err := functions.IDL_DRSBind(rpc, &clientGUID, &clientExt)
	if err != nil {
		rpc.Close()
		return fmt.Errorf("msdrsr: IDL_DRSBind: %w", err)
	}

	if serverExt != nil {
		if ext, perr := serverExt.ParseInt(); perr == nil {
			c.serverExt = ext
		}
	}
	c.rpc = rpc
	c.handle = handle
	c.bound = true
	return nil
}

// Handle returns the drsuapi context handle established by Connect. It is the null
// handle until Connect succeeds.
func (c *Client) Handle() structures.DRS_HANDLE { return c.handle }

// ServerExtensions returns the capability structure the server returned at IDL_DRSBind,
// or nil if not connected or the server returned none. Callers inspect its DwFlags to
// confirm negotiated features (e.g. STRONG_ENCRYPTION before requesting secrets).
func (c *Client) ServerExtensions() *structures.DRS_EXTENSIONS_INT { return c.serverExt }

// SessionKey returns the NTLM session key negotiated for this connection, used to
// decrypt replicated secrets in IDL_DRSGetNCChanges. It is nil until Connect succeeds.
func (c *Client) SessionKey() []byte {
	if c.rpc == nil {
		return nil
	}
	return c.rpc.SessionKey()
}

// RPC exposes the underlying bound DCE/RPC client so replication call wrappers can issue
// drsuapi methods against the established handle. It is nil until Connect succeeds.
func (c *Client) RPC() *dcerpcclient.Client { return c.rpc }

// Close unbinds the drsuapi handle (IDL_DRSUnbind) and tears down the transport. It is
// safe to call on a Client that never connected.
func (c *Client) Close() error {
	if c.rpc == nil {
		return nil
	}
	var unbindErr error
	if c.bound && !c.handle.IsNull() {
		if _, err := functions.IDL_DRSUnbind(c.rpc, c.handle); err != nil {
			unbindErr = fmt.Errorf("msdrsr: IDL_DRSUnbind: %w", err)
		}
	}
	closeErr := c.rpc.Close()
	c.rpc = nil
	c.bound = false
	c.handle = structures.DRS_HANDLE{}
	if unbindErr != nil {
		return unbindErr
	}
	return closeErr
}
