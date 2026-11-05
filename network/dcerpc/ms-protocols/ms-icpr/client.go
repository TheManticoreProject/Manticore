// Package ms_icpr implements the high-level MS-ICPR (ICertPassage Remote Protocol) client
// API over the ICertPassage DCE/RPC interface (91ae6020-9e3c-11cf-8d7c-00aa00c091be v0.0),
// carried over the \cert named pipe on the IPC$ tree.
//
// It is the protocol layer above the interface: callers hold a connected and
// authenticated SMB client and call Request, which binds \cert, issues the single
// CertServerRequest method, and projects the CA's CERTTRANSBLOB replies into a Result.
// This is the enrollment primitive an NTLM relay uses to obtain a certificate as the
// relayed identity (AD CS ESC8 and related), the RPC counterpart to the HTTP/WSTEP
// enrollment path.
//
// ICertPassage is stateless: CertServerRequest is a single self-contained call with an
// explicit binding handle and no context handles to chain, so, like MS-SRVS and unlike
// the handle-chaining MS-SCMR/MS-RRP layers, Client binds a fresh \cert pipe per Request
// (a fault on one cannot desync the next) rather than holding a persistent association. It
// therefore satisfies msproto.Protocol but not msproto.Session.
//
// References:
//   - [MS-ICPR] 3.2.4.1 CertServerRequest
//   - [MS-WCCE] 3.2.1.4.2.1 ICertRequestD::Request (shared disposition semantics); 2.2.2.2 CERTTRANSBLOB
//   - [MS-RPCE] DCE/RPC over SMB named pipes
package ms_icpr

import (
	ICertPassage "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/91ae6020-9e3c-11cf-8d7c-00aa00c091be/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/msproto"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
)

// PipeDialer is the SMB capability MS-ICPR needs: opening a named-pipe transport over an
// established session. It is an alias of msproto.PipeDialer, shared with the other
// named-pipe protocols; network/smb/client.Client satisfies it.
type PipeDialer = msproto.PipeDialer

// Client exposes MS-ICPR enrollment over an established SMB session. The supplied dialer
// MUST be backed by a client that is connected, authenticated, and tree-connected to IPC$
// before any method is called.
type Client struct {
	binder msproto.Binder
}

// compile-time assertion that Client satisfies the shared protocol contract.
var _ msproto.Protocol = (*Client)(nil)

// New returns an MS-ICPR client over the given pipe dialer (typically a
// *network/smb/client.Client).
func New(dialer PipeDialer) *Client {
	return &Client{binder: msproto.NewPipeBinder(dialer, ICertPassage.PipeName)}
}

// Interface reports the DCE/RPC abstract syntax MS-ICPR speaks (ICertPassage v0.0).
func (c *Client) Interface() syntax.SyntaxID { return ICertPassage.SyntaxID() }

// Close releases resources held by the client. MS-ICPR holds no persistent association —
// each Request binds and tears down its own pipe — so there is nothing to release and
// Close is a no-op that always returns nil. It exists to satisfy msproto.Protocol.
func (c *Client) Close() error { return nil }

// bind opens a fresh \cert pipe and binds the ICertPassage abstract syntax over it. The
// returned close function tears the pipe down.
func (c *Client) bind() (*dcerpcclient.Client, func() error, error) {
	return c.binder.Bind(ICertPassage.SyntaxID())
}
