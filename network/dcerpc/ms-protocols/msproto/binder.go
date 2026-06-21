package msproto

import (
	"fmt"
	"time"

	eptiface "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	eptfunctions "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/tcp"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// DefaultTCPTimeout bounds the endpoint-mapper and target TCP dials and reads when a
// TCPBinder is built without an explicit timeout.
const DefaultTCPTimeout = 10 * time.Second

// Binder opens a transport and binds an abstract syntax over it, returning a ready
// DCE/RPC client plus a close function that tears that transport down. It abstracts the
// one step every MS-protocol shares (transport acquisition + bind) over the differing
// transport provenances, so a protocol client can issue calls without embedding any
// dial/bind plumbing of its own.
//
// A stateless client calls Bind once per workflow and invokes the returned closer
// immediately; a session client calls Bind once in Connect and holds the result.
type Binder interface {
	// Bind opens a transport, performs any transport-level authentication the provenance
	// requires, binds the given abstract syntax, and returns the bound client and its
	// closer. The caller owns the close function and must call it to release the transport.
	Bind(s syntax.SyntaxID) (*dcerpcclient.Client, func() error, error)
}

// PipeBinder binds an interface over a DCE/RPC named pipe borrowed from an established SMB
// session. It performs no DCE/RPC-layer authentication: a named pipe inherits the security
// context of the SMB session it rides on.
type PipeBinder struct {
	dialer PipeDialer
	pipe   string
}

// NewPipeBinder returns a Binder that opens the given named pipe over dialer for every
// Bind call. pipe is the IPC$-relative pipe name (e.g. `\srvsvc`, `\winreg`).
func NewPipeBinder(dialer PipeDialer, pipe string) *PipeBinder {
	return &PipeBinder{dialer: dialer, pipe: pipe}
}

// Bind opens a fresh named-pipe transport and binds s over it.
func (b *PipeBinder) Bind(s syntax.SyntaxID) (*dcerpcclient.Client, func() error, error) {
	transport, err := b.dialer.RPCTransport(b.pipe)
	if err != nil {
		return nil, nil, fmt.Errorf("msproto: open pipe %s: %w", b.pipe, err)
	}
	rpc := dcerpcclient.NewClient(transport)
	if err := rpc.Bind(s); err != nil {
		return nil, nil, fmt.Errorf("msproto: bind %s over %s: %w", s.UUID.ToFormatD(), b.pipe, err)
	}
	return rpc, rpc.Close, nil
}

// TCPBinder binds an interface over an owned ncacn_ip_tcp connection. Unlike a borrowed
// pipe, this transport carries no ambient security context, so TCPBinder authenticates at
// the DCE/RPC layer with NTLM at the configured level (packet privacy by default, which
// the secret-bearing interfaces such as drsuapi require). When Port is zero the target
// port is resolved through the endpoint mapper on TCP/135 for the syntax being bound.
type TCPBinder struct {
	Host      string
	Port      int // 0 resolves via the endpoint mapper
	Creds     *credentials.Credentials
	Timeout   time.Duration
	AuthType  uint8 // DCE/RPC auth type; defaults to NTLMSSP
	AuthLevel uint8 // DCE/RPC auth level; defaults to packet privacy (sign + seal)
}

// NewTCPBinder returns a Binder over ncacn_ip_tcp to host authenticating with creds at
// NTLM packet-privacy level. port may be 0 to resolve the target via the endpoint mapper;
// a zero timeout falls back to DefaultTCPTimeout.
func NewTCPBinder(host string, port int, creds *credentials.Credentials, timeout time.Duration) *TCPBinder {
	if timeout == 0 {
		timeout = DefaultTCPTimeout
	}
	return &TCPBinder{
		Host:      host,
		Port:      port,
		Creds:     creds,
		Timeout:   timeout,
		AuthType:  pdu.AuthTypeNTLMSSP,
		AuthLevel: pdu.AuthLevelPktPrivacy,
	}
}

// Bind resolves the endpoint (unless Port is set), dials it over ncacn_ip_tcp,
// authenticates with NTLM at the configured level, and binds s.
func (b *TCPBinder) Bind(s syntax.SyntaxID) (*dcerpcclient.Client, func() error, error) {
	port := b.Port
	if port == 0 {
		resolved, err := b.resolvePort(s)
		if err != nil {
			return nil, nil, fmt.Errorf("msproto: resolve endpoint: %w", err)
		}
		port = resolved
	}

	tr := tcp.New(b.Host, port)
	tr.SetTimeout(b.Timeout)
	rpc := dcerpcclient.NewClient(tr)

	if err := rpc.SetAuth(b.AuthType, b.AuthLevel, b.Creds); err != nil {
		return nil, nil, fmt.Errorf("msproto: configure auth: %w", err)
	}
	if err := rpc.Bind(s); err != nil {
		return nil, nil, fmt.Errorf("msproto: bind %s on %s:%d: %w", s.UUID.ToFormatD(), b.Host, port, err)
	}
	return rpc, rpc.Close, nil
}

// resolvePort asks the endpoint mapper on TCP/135 for the dynamic port the given syntax is
// bound to, returning the first endpoint that carries a TCP port.
func (b *TCPBinder) resolvePort(s syntax.SyntaxID) (int, error) {
	tr := tcp.New(b.Host, tcp.EndpointMapperPort)
	tr.SetTimeout(b.Timeout)
	ept := dcerpcclient.NewClient(tr)
	if err := ept.Bind(eptiface.SyntaxID()); err != nil {
		return 0, fmt.Errorf("bind endpoint mapper: %w", err)
	}
	defer ept.Close()

	eps, err := eptfunctions.Map(ept, s.UUID, s.MajorVersion, s.MinorVersion)
	if err != nil {
		return 0, fmt.Errorf("ept_map: %w", err)
	}
	for _, ep := range eps {
		if ep.Port != 0 {
			return int(ep.Port), nil
		}
	}
	return 0, fmt.Errorf("endpoint mapper returned no TCP endpoint")
}
