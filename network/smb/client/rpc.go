package client

import dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"

// RPCTransport opens pipeName on the current tree and returns it as a DCE/RPC PDU
// transport for the negotiated dialect (SMB1 or SMB2). It is the bridge from the
// version-agnostic SMB client to the DCE/RPC stack: a caller binds it into a
// network/dcerpc/v5/client.Client (which the returned value satisfies as a
// transport.Transport) to drive MS-RPC interfaces such as srvsvc over the IPC$
// tree, without caring which SMB dialect was negotiated.
//
// The current tree must be IPC$ (TreeConnect("IPC$")) before calling.
func (c *Client) RPCTransport(pipeName string) (dcerpctransport.Transport, error) {
	return c.backend.RPCTransport(pipeName)
}
