// Package mssrvs implements high-level MS-SRVS (Server Service Remote Protocol)
// workflows over the srvsvc DCE/RPC interface
// (4b324fc8-1670-01d3-1278-5a47bf6ee188 v3.0), carried over the \srvsvc named pipe on
// the IPC$ tree.
//
// It is the protocol layer above the interface: callers hold a connected and
// authenticated SMB v1 client and call the convenience methods here (ListShares,
// ListSessions, GetServerInfo), which bind to srvsvc, issue the underlying
// NetrShareEnum / NetrSessionEnum / NetrServerGetInfo calls, and project the NDR
// containers and unions into friendly Go structs. The smbclient-ng tool's shares / who
// / info commands are the intended consumers.
//
// References:
//   - [MS-SRVS] 3.1.4.8 NetrShareEnum, 3.1.4.6 NetrSessionEnum, 3.1.4.17 NetrServerGetInfo
//   - [MS-RPCE] DCE/RPC over SMB named pipes
package mssrvs

import (
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/msproto"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
)

// MaxPreferredLength is passed as the preferred maximum reply length of the Netr*Enum
// calls. 0xFFFFFFFF asks the server to return every entry in a single response, so no
// resume-handle paging is needed; this matches common client enumeration usage.
const MaxPreferredLength uint32 = 0xFFFFFFFF

// PipeDialer is the SMB capability MS-SRVS needs: opening a named-pipe transport over an
// established session. It is an alias of msproto.PipeDialer, shared with the other
// named-pipe protocols; network/smb/client.Client satisfies it.
type PipeDialer = msproto.PipeDialer

// Client exposes MS-SRVS workflows over an established SMB session. The supplied dialer
// MUST be backed by a client that is connected, authenticated, and tree-connected to
// IPC$ before any method is called.
//
// MS-SRVS is stateless: srvsvc calls are mutually independent, so the client binds a
// fresh \srvsvc pipe per workflow (a fault on one cannot desync the next) rather than
// holding a persistent association. It therefore satisfies msproto.Protocol but not
// msproto.Session.
type Client struct {
	binder msproto.Binder
}

// compile-time assertion that Client satisfies the shared protocol contract.
var _ msproto.Protocol = (*Client)(nil)

// New returns an MS-SRVS client over the given pipe dialer (typically a
// *network/smb/client.Client).
func New(dialer PipeDialer) *Client {
	return &Client{binder: msproto.NewPipeBinder(dialer, srvsvc.PipeName)}
}

// Interface reports the DCE/RPC abstract syntax MS-SRVS speaks (srvsvc v3.0).
func (c *Client) Interface() syntax.SyntaxID { return srvsvc.SyntaxID() }

// Close releases resources held by the client. MS-SRVS holds no persistent association —
// each workflow binds and tears down its own pipe — so there is nothing to release and
// Close is a no-op that always returns nil. It exists to satisfy msproto.Protocol.
func (c *Client) Close() error { return nil }

// bind opens a fresh \srvsvc pipe and binds the srvsvc abstract syntax over it. srvsvc
// calls are mutually independent, so each workflow runs on its own pipe and bind: a
// fault on one cannot desync the next. The returned close function tears the pipe down.
func (c *Client) bind() (*dcerpcclient.Client, func() error, error) {
	return c.binder.Bind(srvsvc.SyntaxID())
}
