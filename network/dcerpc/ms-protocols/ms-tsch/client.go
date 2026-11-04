// Package ms_tsch implements the high-level MS-TSCH (Task Scheduler Service Remoting
// Protocol) client API over the ITaskSchedulerService DCE/RPC interface
// (86d35949-83c9-4044-b424-db363231fd0c v1.0).
//
// It is the protocol layer above the interface: callers hold a connected and
// authenticated SMB client and drive a Client, which binds the interface once and then
// exposes the task workflow by name (RegisterTask, Run, Delete), plus RegisterRunDelete —
// the scheduled-task command-execution primitive the atexec family and an NTLM relay use.
//
// Transport: ITaskSchedulerService is defined for ncacn_ip_tcp, but Windows also serves
// it over the \atsvc named pipe on the IPC$ tree, which is how it is reached over an
// established SMB session (as atexec does). This client binds that pipe, so it is a
// msproto.Session like MS-SCMR / MS-RRP: the association is held for the client's
// lifetime and torn down by Close. Tasks are addressed by path, not by handle, so no
// handle chaining is involved.
//
// References:
//   - [MS-TSCH] 3.2.5.4.2 SchRpcRegisterTask, 3.2.5.4.13 SchRpcRun, 3.2.5.4.14 SchRpcDelete
//   - [MS-RPCE] DCE/RPC over SMB named pipes
package ms_tsch

import (
	"errors"
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/msproto"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
)

// pipeName is the IPC$-relative named pipe ITaskSchedulerService is served over on an SMB
// session. The interface descriptor leaves PipeName empty because the IDL binds it to
// ncacn_ip_tcp; in practice Windows exposes it on \atsvc, which is what an SMB-carried
// client uses.
const pipeName = `\atsvc`

// ErrNotConnected is returned by methods invoked before a successful Connect.
var ErrNotConnected = errors.New("ms_tsch: not connected; call Connect first")

// PipeDialer is the SMB capability MS-TSCH needs: opening a named-pipe transport over an
// established session. It is an alias of msproto.PipeDialer, shared with the other
// named-pipe protocols; network/smb/client.Client satisfies it.
type PipeDialer = msproto.PipeDialer

// Client is the connected MS-TSCH client. It carries the binder for the session it is
// reached over and, once Connect has run, the single bound \atsvc association that all
// method calls share. It is a msproto.Session: the association is scoped to one pipe.
type Client struct {
	binder    msproto.Binder
	rpc       *dcerpcclient.Client
	closeRPC  func() error
	connected bool
}

// compile-time assertion that Client satisfies the session contract.
var _ msproto.Session = (*Client)(nil)

// New returns an MS-TSCH client over the given pipe dialer (typically a
// *network/smb/client.Client). The dialer MUST be connected, authenticated, and
// tree-connected to IPC$ before Connect is called.
func New(dialer PipeDialer) *Client {
	return &Client{binder: msproto.NewPipeBinder(dialer, pipeName)}
}

// Interface reports the DCE/RPC abstract syntax MS-TSCH speaks (ITaskSchedulerService v1.0).
func (c *Client) Interface() syntax.SyntaxID { return schrpc.SyntaxID() }

// Connect opens the \atsvc pipe and binds the ITaskSchedulerService abstract syntax,
// establishing the single association that all subsequent calls use. It is idempotent.
func (c *Client) Connect() error {
	if c.connected {
		return nil
	}
	rpc, closeRPC, err := c.binder.Bind(schrpc.SyntaxID())
	if err != nil {
		return fmt.Errorf("ms_tsch: connect atsvc: %w", err)
	}
	c.rpc = rpc
	c.closeRPC = closeRPC
	c.connected = true
	return nil
}

// Close tears down the atsvc association. It is safe to call on a client that was never
// connected.
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
