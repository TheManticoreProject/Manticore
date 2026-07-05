package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrServerTransportDelRequest is the [in] parameter set of NetrServerTransportDel:
// the [unique] server name, the level (must be 0), and the single-pointer
// SERVER_TRANSPORT_INFO_0 buffer (a [ref] pointer, so modeled inline).
type netrServerTransportDelRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	Buffer     mssrvs.SERVER_TRANSPORT_INFO_0
}

func (*netrServerTransportDelRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerTransportDel }

// NetrServerTransportDel calls NetrServerTransportDel (opnum 27), unbinding the server
// from a transport protocol ([MS-SRVS] 3.1.4.23).
func NetrServerTransportDel(rpc ndr.Invoker, serverName string, level uint32, buffer mssrvs.SERVER_TRANSPORT_INFO_0) error {
	req := &netrServerTransportDelRequest{
		ServerName: optWStr(serverName),
		Level:      ndr.DWORD(level),
		Buffer:     buffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrServerTransportDel: %w", err)
	}
	if status := uint32(resp.Status); status != srvsvc.NERR_Success {
		return fmt.Errorf("NetrServerTransportDel failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
