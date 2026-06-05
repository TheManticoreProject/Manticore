package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrServerGetInfoRequest is the [in] parameter set of NetrServerGetInfo: the
// [unique] server name and the info level that selects the returned arm.
type netrServerGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
}

func (*netrServerGetInfoRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerGetInfo }

// netrServerGetInfoResponse is the reply: the [out, switch_is(Level)] SERVER_INFO
// union (which carries its own discriminant Tag) and the NET_API_STATUS.
type netrServerGetInfoResponse struct {
	InfoStruct structures.SERVER_INFO
	Status     ndr.DWORD `ndr:"retval"`
}

// NetrServerGetInfo calls NetrServerGetInfo (opnum 21), retrieving configuration
// information for the specified server ([MS-SRVS] 3.1.4.17).
func NetrServerGetInfo(rpc ndr.Invoker, serverName string, level uint32) (structures.SERVER_INFO, error) {
	req := &netrServerGetInfoRequest{
		ServerName: optWStr(serverName),
		Level:      ndr.DWORD(level),
	}
	var resp netrServerGetInfoResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return structures.SERVER_INFO{}, fmt.Errorf("NetrServerGetInfo: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, fmt.Errorf("NetrServerGetInfo failed: %s", srvsvc.StatusString(status))
	}
	return resp.InfoStruct, nil
}
