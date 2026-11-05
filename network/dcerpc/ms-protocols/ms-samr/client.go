// Package ms_samr implements the high-level MS-SAMR (Security Account Manager Remote
// Protocol) client API over the samr DCE/RPC interface
// (12345778-1234-abcd-ef00-0123456789ac v1.0), carried over the \samr named pipe on the
// IPC$ tree.
//
// It is the protocol layer above the interface: callers hold a connected and
// authenticated SMB client and drive a Client, which binds samr once and then exposes the
// account-enumeration workflow by name (OpenServer, OpenDomain, LookupDomain, OpenAlias,
// GetMembersInAlias, OpenGroup, GetMembersInGroup, CloseHandle), plus LocalAliasMembers —
// the enum-local-admins primitive an NTLM relay or a pentest reaches for.
//
// Like MS-RRP and MS-SCMR and unlike the stateless MS-SRVS layer, SAM context handles
// CHAIN (server -> domain -> alias/group -> close) and are scoped to the RPC association,
// so Client binds a SINGLE \samr pipe in Connect and reuses it for the client's lifetime;
// a handle obtained on one association cannot be used on another. Close tears the pipe
// down, releasing any server-side handles.
//
// It is scoped to enumeration. Account creation (SamrCreateUser2InDomain, for
// add-computer) is intentionally left as a follow-up.
//
// References:
//   - [MS-SAMR] 3.1.5 samr method behaviours; 2.2.3 handles; 2.2.1.1 access masks
//   - [MS-RPCE] DCE/RPC over SMB named pipes
package ms_samr

import (
	"errors"
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/msproto"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// Handle is an open SAM context handle (server, domain, alias, group or user) returned by
// the Open* methods. It aliases the interface's SAMPR_HANDLE so it can be passed back to
// any method without conversion.
type Handle = mssamr.SAMPR_HANDLE

// ErrNotConnected is returned by methods invoked before a successful Connect.
var ErrNotConnected = errors.New("ms_samr: not connected; call Connect first")

// PipeDialer is the SMB capability MS-SAMR needs: opening a named-pipe transport over an
// established session. It is an alias of msproto.PipeDialer, shared with the other
// named-pipe protocols; network/smb/client.Client satisfies it.
type PipeDialer = msproto.PipeDialer

// Client is the connected MS-SAMR client. It carries the binder for the session it is
// reached over and, once Connect has run, the single bound \samr association that all
// method calls and chained handles share. It is a msproto.Session: SAM handles chain and
// are scoped to that one association.
type Client struct {
	binder    msproto.Binder
	rpc       *dcerpcclient.Client
	closeRPC  func() error
	connected bool
}

// compile-time assertion that Client satisfies the session contract.
var _ msproto.Session = (*Client)(nil)

// New returns an MS-SAMR client over the given pipe dialer (typically a
// *network/smb/client.Client). The dialer MUST be connected, authenticated, and
// tree-connected to IPC$ before Connect is called.
func New(dialer PipeDialer) *Client {
	return &Client{binder: msproto.NewPipeBinder(dialer, samr.PipeName)}
}

// Interface reports the DCE/RPC abstract syntax MS-SAMR speaks (samr v1.0).
func (c *Client) Interface() syntax.SyntaxID { return samr.SyntaxID() }

// Connect opens the \samr pipe and binds the samr abstract syntax, establishing the
// single association that all subsequent calls and handles use. It is idempotent.
func (c *Client) Connect() error {
	if c.connected {
		return nil
	}
	rpc, closeRPC, err := c.binder.Bind(samr.SyntaxID())
	if err != nil {
		return fmt.Errorf("ms_samr: connect samr: %w", err)
	}
	c.rpc = rpc
	c.closeRPC = closeRPC
	c.connected = true
	return nil
}

// Close tears down the samr association, releasing any server-side handles still open on
// it. It is safe to call on a client that was never connected.
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
