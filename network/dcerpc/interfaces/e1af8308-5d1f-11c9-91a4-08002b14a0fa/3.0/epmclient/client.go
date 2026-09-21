// Package epmclient provides the high-level endpoint mapper (ept) client, wrapping the
// transport setup, bind, and call flow for both named-pipe (ncacn_np) and TCP (ncacn_ip_tcp)
// transports, mirroring the MS-RRP client pattern.
//
// Usage over SMB (ncacn_np):
//
//	epm := epmclient.New(smbClient)
//	epm.Connect()
//	defer epm.Close()
//	entries, _ := epm.Lookup()
//
// Usage over TCP:
//
//	epm := epmclient.NewOverTCP("10.0.0.1", 135, 0)
//	epm.Connect()
//	defer epm.Close()
//	eps, _ := epm.Map(interfaceUUID, 1, 0)
package epmclient

import (
	"errors"
	"fmt"
	"time"

	eptiface "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/msproto"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/tcp"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msrpce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpce"
)

// PipeDialer is the SMB capability the endpoint mapper client needs when reached over a
// named pipe: opening a named-pipe transport over an established session. It is an alias
// of msproto.PipeDialer; network/smb/client.Client satisfies it.
type PipeDialer = msproto.PipeDialer

// ErrNotConnected is returned by methods invoked before a successful Connect.
var ErrNotConnected = errors.New("epm: not connected; call Connect first")

// defaultTimeout bounds TCP dials and reads when no explicit timeout is supplied.
const defaultTimeout = 10 * time.Second

// EndpointMapper is the high-level endpoint mapper (ept) client. It wraps the transport
// setup, bind, and call flow so callers never touch the UUID-named interface package or
// wire plumbing directly.
type EndpointMapper struct {
	binder    msproto.Binder
	rpc       *dcerpcclient.Client
	closeRPC  func() error
	connected bool
}

var _ msproto.Session = (*EndpointMapper)(nil)

// New returns an EndpointMapper over the given pipe dialer (typically a
// *network/smb/client.Client). The dialer must be connected, authenticated, and
// tree-connected to IPC$ before Connect is called.
func New(dialer PipeDialer) *EndpointMapper {
	return &EndpointMapper{binder: msproto.NewPipeBinder(dialer, eptiface.PipeName)}
}

// NewOverTCP returns an EndpointMapper that connects directly over ncacn_ip_tcp. port is
// typically 135 (tcp.EndpointMapperPort); zero defaults to 135. A zero timeout falls back
// to 10 seconds.
func NewOverTCP(host string, port int, timeout time.Duration) *EndpointMapper {
	if port == 0 {
		port = tcp.EndpointMapperPort
	}
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &EndpointMapper{binder: &tcpBinder{host: host, port: port, timeout: timeout}}
}

// Interface reports the DCE/RPC abstract syntax the endpoint mapper speaks (ept v3.0).
func (e *EndpointMapper) Interface() syntax.SyntaxID { return eptiface.SyntaxID() }

// Connect opens the transport and binds the ept abstract syntax, establishing the
// association that all subsequent calls use. It is idempotent.
func (e *EndpointMapper) Connect() error {
	if e.connected {
		return nil
	}
	rpc, closeRPC, err := e.binder.Bind(eptiface.SyntaxID())
	if err != nil {
		return fmt.Errorf("epm: connect: %w", err)
	}
	e.rpc = rpc
	e.closeRPC = closeRPC
	e.connected = true
	return nil
}

// Close tears down the ept association. It is safe to call on a client that was never
// connected.
func (e *EndpointMapper) Close() error {
	if !e.connected || e.closeRPC == nil {
		return nil
	}
	err := e.closeRPC()
	e.rpc = nil
	e.closeRPC = nil
	e.connected = false
	return err
}

// IsConnected reports whether Connect has succeeded and Close has not yet run.
func (e *EndpointMapper) IsConnected() bool { return e.connected && e.rpc != nil }

func (e *EndpointMapper) ensure() error {
	if !e.connected || e.rpc == nil {
		return ErrNotConnected
	}
	return nil
}

// Lookup enumerates the entire endpoint map by paging ept_lookup to completion. It
// returns every ept_entry_t the server holds. For finer control (filtering by interface
// or object UUID) call the low-level functions.EptLookup directly on the RPC client
// obtained via Connect.
func (e *EndpointMapper) Lookup() ([]msrpce.EptEntry, error) {
	if err := e.ensure(); err != nil {
		return nil, err
	}
	return functions.Lookup(e.rpc)
}

// Map resolves the ncacn_ip_tcp endpoints bound to the given interface UUID and version.
// It builds a TCP map tower and calls ept_map, then extracts the endpoints from the
// returned towers. For a non-TCP tower or finer control, call the low-level
// functions.EptMap directly.
func (e *EndpointMapper) Map(iface guid.GUID, major, minor uint16) ([]msrpce.Endpoint, error) {
	if err := e.ensure(); err != nil {
		return nil, err
	}
	return functions.Map(e.rpc, iface, major, minor)
}

// tcpBinder implements msproto.Binder for unauthenticated TCP connections to the endpoint
// mapper. EPM on TCP/135 requires no DCE/RPC-level authentication.
type tcpBinder struct {
	host    string
	port    int
	timeout time.Duration
}

func (b *tcpBinder) Bind(s syntax.SyntaxID) (*dcerpcclient.Client, func() error, error) {
	tr := tcp.New(b.host, b.port)
	tr.SetTimeout(b.timeout)
	rpc := dcerpcclient.NewClient(tr)
	if err := rpc.Bind(s); err != nil {
		return nil, nil, fmt.Errorf("epm: bind %s on %s:%d: %w", s.UUID.ToFormatD(), b.host, b.port, err)
	}
	return rpc, rpc.Close, nil
}
