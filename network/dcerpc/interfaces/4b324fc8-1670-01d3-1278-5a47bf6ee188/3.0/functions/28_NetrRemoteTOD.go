package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrRemoteTODRequest is the [in] parameter set of NetrRemoteTOD: just the [unique]
// server name.
type netrRemoteTODRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
}

func (*netrRemoteTODRequest) Opnum() uint16 { return srvsvc.OpnumNetrRemoteTOD }

// netrRemoteTODResponse is the reply: the [out] double pointer to a TIME_OF_DAY_INFO
// (a [unique] referent) and the NET_API_STATUS return value.
type netrRemoteTODResponse struct {
	BufferPtr *structures.TIME_OF_DAY_INFO `ndr:"unique"`
	Status    ndr.DWORD                    `ndr:"retval"`
}

// NetrRemoteTOD calls NetrRemoteTOD (opnum 28), retrieving the time of day on the
// server ([MS-SRVS] 3.1.4.24).
func NetrRemoteTOD(rpc ndr.Invoker, serverName string) (*structures.TIME_OF_DAY_INFO, error) {
	req := &netrRemoteTODRequest{
		ServerName: optWStr(serverName),
	}
	var resp netrRemoteTODResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("NetrRemoteTOD: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success {
		return resp.BufferPtr, fmt.Errorf("NetrRemoteTOD failed: %s", srvsvc.StatusString(status))
	}
	return resp.BufferPtr, nil
}
