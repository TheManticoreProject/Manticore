// Package ms_rrp implements the high-level MS-RRP (Windows Remote Registry Protocol)
// client API over the winreg DCE/RPC interface
// (338cd001-2244-31f1-aaaa-900038001003 v1.0), carried over the \winreg named pipe on
// the IPC$ tree.
//
// It is the protocol layer above the interface: callers hold a connected and
// authenticated SMB client and drive a RemoteRegistry, which binds winreg once and then
// exposes every interface method by its exact name (OpenLocalMachine, BaseRegOpenKey,
// BaseRegQueryValue, ...) as a method on the struct, plus a small set of ergonomic,
// reg.exe-style helpers (OpenKeyByPath, QueryValueByPath, EnumKeys, ...).
//
// Unlike the stateless MS-SRVS layer, registry context handles CHAIN
// (HKLM -> subkey -> close) and are scoped to the RPC association, so RemoteRegistry
// binds a SINGLE \winreg pipe in Connect and reuses it for the lifetime of the client; a
// handle obtained on one association cannot be used on another. Close tears the pipe
// down, which releases any server-side handles.
//
// References:
//   - [MS-RRP] 3.1.5 winreg method behaviours; 2.2.4 REGSAM; [MS-ERREF] 2.2 Win32 codes
//   - [MS-RPCE] DCE/RPC over SMB named pipes
package ms_rrp

import (
	"errors"
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
)

// Handle is an open registry key context handle returned by the Open*/BaseRegCreateKey/
// BaseRegOpenKey methods. It aliases the interface's RPC_HKEY so it can be passed back to
// any method without conversion.
type Handle = structures.RPC_HKEY

// ErrNotConnected is returned by methods invoked before a successful Connect.
var ErrNotConnected = errors.New("ms_rrp: not connected; call Connect first")

// PipeDialer opens a fresh DCE/RPC named-pipe transport over an established, IPC$-tree-
// connected SMB session. It is the only capability MS-RRP needs from the SMB layer, so
// depending on this interface (rather than a concrete SMB client) keeps ms_rrp
// independent of the SMB dialect: network/smb/client.Client satisfies it (via its
// RPCTransport method) and routes to an SMB1 or SMB2 named-pipe transport according to
// what was negotiated.
type PipeDialer interface {
	RPCTransport(pipeName string) (dcerpctransport.Transport, error)
}

// RemoteRegistry is the connected MS-RRP client. It carries the session it is reached
// over (the pipe dialer) and, once Connect has run, the single bound \winreg association
// that all method calls and chained key handles share.
type RemoteRegistry struct {
	dialer    PipeDialer
	rpc       *dcerpcclient.Client
	closeRPC  func() error
	connected bool
}

// New returns a RemoteRegistry over the given pipe dialer (typically a
// *network/smb/client.Client). The dialer MUST be connected, authenticated, and
// tree-connected to IPC$ before Connect is called.
func New(dialer PipeDialer) *RemoteRegistry {
	return &RemoteRegistry{dialer: dialer}
}

// Connect opens the \winreg pipe and binds the winreg abstract syntax, establishing the
// single association that all subsequent calls and key handles use. It is idempotent:
// calling it on an already-connected client is a no-op.
func (r *RemoteRegistry) Connect() error {
	if r.connected {
		return nil
	}
	transport, err := r.dialer.RPCTransport(winreg.PipeName)
	if err != nil {
		return fmt.Errorf("ms_rrp: open winreg pipe: %w", err)
	}
	rpc := dcerpcclient.NewClient(transport)
	if err := rpc.Bind(winreg.SyntaxID()); err != nil {
		return fmt.Errorf("ms_rrp: bind winreg: %w", err)
	}
	r.rpc = rpc
	r.closeRPC = rpc.Close
	r.connected = true
	return nil
}

// Close tears down the winreg association, releasing any server-side key handles still
// open on it. It is safe to call on a client that was never connected.
func (r *RemoteRegistry) Close() error {
	if !r.connected || r.closeRPC == nil {
		return nil
	}
	err := r.closeRPC()
	r.rpc = nil
	r.closeRPC = nil
	r.connected = false
	return err
}

// IsConnected reports whether Connect has succeeded and Close has not yet run.
func (r *RemoteRegistry) IsConnected() bool { return r.connected && r.rpc != nil }

// ensure guards method calls against use before Connect.
func (r *RemoteRegistry) ensure() error {
	if !r.connected || r.rpc == nil {
		return ErrNotConnected
	}
	return nil
}
