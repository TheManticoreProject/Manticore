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

	eptiface "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	eptfunctions "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/tcp"
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
	bound     bool
}

// New returns an MS-DRSR client for the given host (IP or hostname) and credentials. By
// default the drsuapi TCP endpoint is resolved through the endpoint mapper on TCP/135;
// call SetPort to skip resolution and dial a known port directly.
func New(host string, creds *credentials.Credentials) *Client {
	return &Client{host: host, creds: creds, timeout: DefaultTimeout}
}

// SetTimeout bounds each TCP dial and read. It must be called before Connect.
func (c *Client) SetTimeout(d time.Duration) { c.timeout = d }

// SetPort overrides endpoint-mapper resolution and dials drsuapi on the given TCP port
// directly. It must be called before Connect.
func (c *Client) SetPort(port int) { c.port = port }

// resolvePort asks the endpoint mapper on TCP/135 for the dynamic port the drsuapi
// interface is bound to, returning the first endpoint that carries a TCP port.
func (c *Client) resolvePort() (int, error) {
	tr := tcp.New(c.host, tcp.EndpointMapperPort)
	tr.SetTimeout(c.timeout)
	ept := dcerpcclient.NewClient(tr)
	if err := ept.Bind(eptiface.SyntaxID()); err != nil {
		return 0, fmt.Errorf("bind endpoint mapper: %w", err)
	}
	defer ept.Close()

	syntax := drsuapi.SyntaxID()
	eps, err := eptfunctions.Map(ept, syntax.UUID, syntax.MajorVersion, syntax.MinorVersion)
	if err != nil {
		return 0, fmt.Errorf("ept_map drsuapi: %w", err)
	}
	for _, ep := range eps {
		if ep.Port != 0 {
			return int(ep.Port), nil
		}
	}
	return 0, fmt.Errorf("endpoint mapper returned no TCP endpoint for drsuapi")
}

// Connect resolves the drsuapi endpoint (unless a port was set), dials it over
// ncacn_ip_tcp, authenticates with NTLM at packet-privacy level, binds the drsuapi
// abstract syntax, and performs the IDL_DRSBind capability handshake. On success the
// negotiated context handle and the server's extensions are available via Handle and
// ServerExtensions, and the connection is ready for replication calls.
func (c *Client) Connect() error {
	if c.bound {
		return fmt.Errorf("msdrsr: already connected")
	}

	port := c.port
	if port == 0 {
		resolved, err := c.resolvePort()
		if err != nil {
			return fmt.Errorf("msdrsr: resolve drsuapi endpoint: %w", err)
		}
		port = resolved
	}

	tr := tcp.New(c.host, port)
	tr.SetTimeout(c.timeout)
	rpc := dcerpcclient.NewClient(tr)

	// drsuapi requires packet privacy (sign + seal); the server rejects lower levels.
	if err := rpc.SetAuth(pdu.AuthTypeNTLMSSP, pdu.AuthLevelPktPrivacy, c.creds); err != nil {
		return fmt.Errorf("msdrsr: configure auth: %w", err)
	}
	if err := rpc.Bind(drsuapi.SyntaxID()); err != nil {
		return fmt.Errorf("msdrsr: bind drsuapi on %s:%d: %w", c.host, port, err)
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
