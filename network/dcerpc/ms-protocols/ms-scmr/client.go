// Package ms_scmr implements the high-level MS-SCMR (Service Control Manager Remote
// Protocol) client API over the svcctl DCE/RPC interface
// (367abb81-9844-35f1-ad32-98f038001003 v2.0), carried over the \svcctl named pipe on the
// IPC$ tree.
//
// It is the protocol layer above the interface: callers hold a connected and
// authenticated SMB client and drive a Client, which binds svcctl once and then exposes
// the service-control workflow by name (OpenSCManager, CreateService, StartService,
// ControlService, DeleteService, CloseHandle), plus CreateAndRun — the service-based
// command-execution primitive the psexec/smbexec family and an NTLM relay use.
//
// Like MS-RRP and unlike the stateless MS-SRVS layer, SCM context handles CHAIN (SCM
// handle -> service handle -> close) and are scoped to the RPC association, so Client
// binds a SINGLE \svcctl pipe in Connect and reuses it for the client's lifetime; a
// handle obtained on one association cannot be used on another. Close tears the pipe
// down, releasing any server-side handles.
//
// References:
//   - [MS-SCMR] 3.1.4 svcctl method behaviours; 2.2.4 handles; 2.2.15 service access rights
//   - [MS-RPCE] DCE/RPC over SMB named pipes
package ms_scmr

import (
	"errors"
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/msproto"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// Handle is an open SCM or service context handle returned by OpenSCManager, OpenService
// and CreateService. It aliases the interface's SC_RPC_HANDLE so it can be passed back to
// any method without conversion.
type Handle = msscmr.SC_RPC_HANDLE

// ErrNotConnected is returned by methods invoked before a successful Connect.
var ErrNotConnected = errors.New("ms_scmr: not connected; call Connect first")

// PipeDialer is the SMB capability MS-SCMR needs: opening a named-pipe transport over an
// established session. It is an alias of msproto.PipeDialer, shared with the other
// named-pipe protocols; network/smb/client.Client satisfies it.
type PipeDialer = msproto.PipeDialer

// Client is the connected MS-SCMR client. It carries the binder for the session it is
// reached over and, once Connect has run, the single bound \svcctl association that all
// method calls and chained handles share. It is a msproto.Session: SCM handles chain and
// are scoped to that one association.
type Client struct {
	binder    msproto.Binder
	rpc       *dcerpcclient.Client
	closeRPC  func() error
	connected bool
}

// compile-time assertion that Client satisfies the session contract.
var _ msproto.Session = (*Client)(nil)

// New returns an MS-SCMR client over the given pipe dialer (typically a
// *network/smb/client.Client). The dialer MUST be connected, authenticated, and
// tree-connected to IPC$ before Connect is called.
func New(dialer PipeDialer) *Client {
	return &Client{binder: msproto.NewPipeBinder(dialer, svcctl.PipeName)}
}

// Interface reports the DCE/RPC abstract syntax MS-SCMR speaks (svcctl v2.0).
func (c *Client) Interface() syntax.SyntaxID { return svcctl.SyntaxID() }

// Connect opens the \svcctl pipe and binds the svcctl abstract syntax, establishing the
// single association that all subsequent calls and handles use. It is idempotent.
func (c *Client) Connect() error {
	if c.connected {
		return nil
	}
	rpc, closeRPC, err := c.binder.Bind(svcctl.SyntaxID())
	if err != nil {
		return fmt.Errorf("ms_scmr: connect svcctl: %w", err)
	}
	c.rpc = rpc
	c.closeRPC = closeRPC
	c.connected = true
	return nil
}

// Close tears down the svcctl association, releasing any server-side handles still open
// on it. It is safe to call on a client that was never connected.
func (c *Client) Close() error {
	if !c.connected || c.closeRPC == nil {
		return nil
	}
	err := c.closeRPC()
	c.rpc = nil
	c.closeRPC = nil
	c.connected = false
	return err
}

// IsConnected reports whether Connect has succeeded and Close has not yet run.
func (c *Client) IsConnected() bool { return c.connected && c.rpc != nil }

// ensure guards method calls against use before Connect.
func (c *Client) ensure() error {
	if !c.connected || c.rpc == nil {
		return ErrNotConnected
	}
	return nil
}
