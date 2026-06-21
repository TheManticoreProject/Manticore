// Package msproto holds the shared contracts and transport-binding helpers common to
// every high-level MS-protocol client (ms-srvs, ms-rrp, ms-drsr, ...).
//
// An MS-protocol client is the workflow layer above a single DCE/RPC interface: it binds
// that interface's abstract syntax over some transport and exposes friendly methods. The
// clients differ on two orthogonal axes — where the transport comes from (a borrowed SMB
// named pipe vs. an owned ncacn_ip_tcp connection) and how long a bound association lives
// (a persistent bind reused across calls vs. a fresh bind per call). This package captures
// only what is genuinely common across those axes:
//
//   - Protocol: every client reports the abstract syntax it speaks and can be Closed.
//   - Session:  the subset of clients that hold a persistent bound association also expose
//     Connect/IsConnected.
//   - Binder:   the "open a transport and bind a syntax" step, with one implementation per
//     transport provenance (PipeBinder over SMB, TCPBinder over ncacn_ip_tcp).
//
// It deliberately does NOT unify context-handle types (those are interface-specific) nor
// the bind-per-call vs. persistent-bind policy (that stays each client's choice).
//
// This package depends only on the DCE/RPC transport, client, syntax, and credentials
// layers; the MS-protocol packages depend on it, never the reverse.
package msproto

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
)

// Protocol is the contract every MS-protocol client satisfies, regardless of transport
// provenance or session model.
type Protocol interface {
	// Interface reports the DCE/RPC abstract syntax (UUID + version) this protocol speaks.
	Interface() syntax.SyntaxID

	// Close releases everything the client holds — server-side context handles and any
	// owned transport. It follows io.Closer semantics: it is safe to call when nothing is
	// held (e.g. a stateless client, or one that never connected) and returns nil then.
	Close() error
}

// Session is implemented by the MS-protocol clients that hold a persistent bound
// association for the lifetime of the client (their context handles chain across calls and
// are scoped to that one association). Stateless clients that bind a fresh transport per
// call satisfy Protocol but not Session.
type Session interface {
	Protocol

	// Connect establishes the persistent association (resolving endpoints and binding the
	// abstract syntax as the protocol requires). It is idempotent: calling it on an
	// already-connected client is a no-op that returns nil.
	Connect() error

	// IsConnected reports whether Connect has succeeded and Close has not yet run.
	IsConnected() bool
}

// PipeDialer opens a fresh DCE/RPC named-pipe transport for the given pipe over an
// established, IPC$-tree-connected SMB session. It is the only capability the named-pipe
// MS-protocols (ms-srvs, ms-rrp) need from the SMB layer, so depending on this interface
// rather than a concrete SMB client keeps them independent of the SMB dialect:
// network/smb/client.Client satisfies it (via its RPCTransport method) and routes to an
// SMB1 or SMB2 named-pipe transport according to what was negotiated.
type PipeDialer interface {
	RPCTransport(pipeName string) (dcerpctransport.Transport, error)
}
