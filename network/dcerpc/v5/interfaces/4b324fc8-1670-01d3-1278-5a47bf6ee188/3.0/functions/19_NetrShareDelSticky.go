package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
)

// netrShareDelStickyRequest is the [in] parameter set of NetrShareDelSticky: the optional
// server name, the [in,string] (ref) share name, and a reserved DWORD (MUST be zero).
type netrShareDelStickyRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	NetName    ndr.WSTR
	Reserved   ndr.DWORD
}

func (*netrShareDelStickyRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareDelSticky
}

// NetrShareDelSticky calls NetrShareDelSticky (opnum 19), marking a sticky share as
// non-persistent so it is not recreated at restart ([MS-SRVS] 3.1.4.13).
func NetrShareDelSticky(rpc *client.Client, serverName string, netName string, reserved ndr.DWORD) error {
	req := &netrShareDelStickyRequest{
		ServerName: optWStr(serverName),
		NetName:    ndr.WSTR(netName),
		Reserved:   reserved,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrShareDelSticky: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrShareDelSticky failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return nil
}
