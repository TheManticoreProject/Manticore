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
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
)

// MaxPreferredLength is passed as the preferred maximum reply length of the Netr*Enum
// calls. 0xFFFFFFFF asks the server to return every entry in a single response, so no
// resume-handle paging is needed; this matches common client enumeration usage.
const MaxPreferredLength uint32 = 0xFFFFFFFF

// PipeDialer opens a fresh DCE/RPC named-pipe transport for the given pipe over an
// established, IPC$-tree-connected SMB session. It is the only capability MS-SRVS needs
// from the SMB layer, so depending on this interface (rather than a concrete SMB
// client) keeps mssrvs independent of the SMB dialect: the generic
// network/smb/client.Client satisfies it (via its RPCTransport method) and routes to an
// SMB1 or SMB2 named-pipe transport according to what was negotiated.
type PipeDialer interface {
	RPCTransport(pipeName string) (dcerpctransport.Transport, error)
}

// Client exposes MS-SRVS workflows over an established SMB session. The supplied dialer
// MUST be backed by a client that is connected, authenticated, and tree-connected to
// IPC$ before any method is called.
type Client struct {
	dialer PipeDialer
}

// New returns an MS-SRVS client over the given pipe dialer (typically a
// *network/smb/client.Client).
func New(dialer PipeDialer) *Client {
	return &Client{dialer: dialer}
}

// bind opens a fresh \srvsvc pipe and binds the srvsvc abstract syntax over it. srvsvc
// calls are mutually independent, so each workflow runs on its own pipe and bind: a
// fault on one cannot desync the next. The returned close function tears the pipe down.
func (c *Client) bind() (*dcerpcclient.Client, func() error, error) {
	transport, err := c.dialer.RPCTransport(srvsvc.PipeName)
	if err != nil {
		return nil, nil, fmt.Errorf("mssrvs: open srvsvc pipe: %w", err)
	}
	rpc := dcerpcclient.NewClient(transport)
	if err := rpc.Bind(srvsvc.SyntaxID()); err != nil {
		return nil, nil, fmt.Errorf("mssrvs: bind srvsvc: %w", err)
	}
	return rpc, rpc.Close, nil
}
