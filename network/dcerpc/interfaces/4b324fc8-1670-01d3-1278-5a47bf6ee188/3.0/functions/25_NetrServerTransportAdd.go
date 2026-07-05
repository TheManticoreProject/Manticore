package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrServerTransportAddRequest is the [in] parameter set of NetrServerTransportAdd:
// the [unique] server name, the level (must be 0), and the single-pointer
// SERVER_TRANSPORT_INFO_0 buffer (a [ref] pointer, so modeled inline).
type netrServerTransportAddRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	Buffer     mssrvs.SERVER_TRANSPORT_INFO_0
}

func (*netrServerTransportAddRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerTransportAdd }

// NetrServerTransportAdd calls NetrServerTransportAdd (opnum 25), binding the server
// to a transport protocol ([MS-SRVS] 3.1.4.21).
func NetrServerTransportAdd(rpc ndr.Invoker, serverName string, level uint32, buffer mssrvs.SERVER_TRANSPORT_INFO_0) error {
	req := &netrServerTransportAddRequest{
		ServerName: optWStr(serverName),
		Level:      ndr.DWORD(level),
		Buffer:     buffer,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrServerTransportAdd: %w", err)
	}
	if status := uint32(resp.Status); status != srvsvc.NERR_Success {
		return fmt.Errorf("NetrServerTransportAdd failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
