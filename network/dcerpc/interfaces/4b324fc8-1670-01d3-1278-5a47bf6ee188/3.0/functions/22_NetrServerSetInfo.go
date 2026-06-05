package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrServerSetInfoRequest is the [in]/[in,out] parameter set of NetrServerSetInfo:
// the [unique] server name, the info level, the [in, switch_is(Level)] SERVER_INFO
// union (its Tag is set to Level before marshalling), and the [in,out,unique]
// parameter-in-error index.
type netrServerSetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	ServerInfo structures.SERVER_INFO
	ParmErr    *ndr.DWORD `ndr:"unique"`
}

func (*netrServerSetInfoRequest) Opnum() uint16 { return srvsvc.OpnumNetrServerSetInfo }

// netrServerSetInfoResponse is the reply: the updated [in,out,unique] parameter-in-error
// index and the NET_API_STATUS return value.
type netrServerSetInfoResponse struct {
	ParmErr *ndr.DWORD `ndr:"unique"`
	Status  ndr.DWORD  `ndr:"retval"`
}

// NetrServerSetInfo calls NetrServerSetInfo (opnum 22), setting the server's
// configuration parameters ([MS-SRVS] 3.1.4.18). On ERROR_INVALID_PARAMETER the
// returned parameter-error index identifies the offending field.
func NetrServerSetInfo(rpc ndr.Invoker, serverName string, level uint32, serverInfo structures.SERVER_INFO, parmErr uint32) (uint32, error) {
	serverInfo.Tag = ndr.DWORD(level)
	pe := ndr.DWORD(parmErr)
	req := &netrServerSetInfoRequest{
		ServerName: optWStr(serverName),
		Level:      ndr.DWORD(level),
		ServerInfo: serverInfo,
		ParmErr:    &pe,
	}
	var resp netrServerSetInfoResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return parmErr, fmt.Errorf("NetrServerSetInfo: %w", err)
	}
	var outParmErr uint32
	if resp.ParmErr != nil {
		outParmErr = uint32(*resp.ParmErr)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success {
		return outParmErr, fmt.Errorf("NetrServerSetInfo failed: %s", srvsvc.StatusString(status))
	}
	return outParmErr, nil
}
