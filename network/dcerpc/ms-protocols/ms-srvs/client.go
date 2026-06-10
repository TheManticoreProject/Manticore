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
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
)

// MaxPreferredLength is passed as the preferred maximum reply length of the Netr*Enum
// calls. 0xFFFFFFFF asks the server to return every entry in a single response, so no
// resume-handle paging is needed; this matches impacket's enumeration usage.
const MaxPreferredLength uint32 = 0xFFFFFFFF

// Client exposes MS-SRVS workflows over an established SMB v1 session. The supplied SMB
// client MUST already be connected, have an authenticated session (SessionSetup), and
// have a tree connected to IPC$ (TreeConnect) before any method is called.
type Client struct {
	smb *smbclient.Client
}

// New returns an MS-SRVS client over the given SMB v1 client.
func New(smb *smbclient.Client) *Client {
	return &Client{smb: smb}
}

// bind opens a fresh \srvsvc pipe and binds the srvsvc abstract syntax over it. srvsvc
// calls are mutually independent, so each workflow runs on its own pipe and bind: a
// fault on one cannot desync the next. The returned close function tears the pipe down.
func (c *Client) bind() (*dcerpcclient.Client, func() error, error) {
	rpc := dcerpcclient.NewClient(dcerpcsmb.New(c.smb, srvsvc.PipeName))
	if err := rpc.Bind(srvsvc.SyntaxID()); err != nil {
		return nil, nil, fmt.Errorf("mssrvs: bind srvsvc: %w", err)
	}
	return rpc, rpc.Close, nil
}
