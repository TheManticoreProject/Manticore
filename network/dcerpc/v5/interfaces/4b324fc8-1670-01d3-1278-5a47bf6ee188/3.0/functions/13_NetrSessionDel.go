package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
)

// netrSessionDelRequest is the [in] parameter set of NetrSessionDel: the [unique] server
// name and the [unique] client-name and user-name selecting the session(s) to end.
type netrSessionDelRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	ClientName *ndr.WSTR `ndr:"unique"`
	UserName   *ndr.WSTR `ndr:"unique"`
}

func (*netrSessionDelRequest) Opnum() uint16 { return srvsvc.OpnumNetrSessionDel }

// NetrSessionDel calls NetrSessionDel (opnum 13), ending the network session(s) between
// the server and the named client/user ([MS-SRVS] 3.1.4.6).
func NetrSessionDel(rpc *client.Client, serverName, clientName, userName string) error {
	req := &netrSessionDelRequest{
		ServerName: optWStr(serverName),
		ClientName: optWStr(clientName),
		UserName:   optWStr(userName),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrSessionDel: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrSessionDel failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
