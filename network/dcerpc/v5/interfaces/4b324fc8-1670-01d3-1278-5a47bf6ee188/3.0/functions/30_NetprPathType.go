package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
)

// netprPathTypeRequest is the [in] parameter set of NetprPathType: the [unique] server
// name, the [in,string] (ref) path name, and the type flags.
type netprPathTypeRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	PathName   ndr.WSTR
	Flags      ndr.DWORD
}

func (*netprPathTypeRequest) Opnum() uint16 { return srvsvc.OpnumNetprPathType }

// netprPathTypeResponse is the reply: the [out] PathType and the NET_API_STATUS return
// value.
type netprPathTypeResponse struct {
	PathType ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// NetprPathType calls NetprPathType (opnum 30), checking a path name's syntactic type
// ([MS-SRVS] 3.1.4.26).
func NetprPathType(rpc *client.Client, serverName, pathName string, flags uint32) (uint32, error) {
	req := &netprPathTypeRequest{
		ServerName: optWStr(serverName),
		PathName:   ndr.WSTR(pathName),
		Flags:      ndr.DWORD(flags),
	}
	var resp netprPathTypeResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("NetprPathType: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return uint32(resp.PathType), fmt.Errorf("NetprPathType failed: %s", srvsvc.StatusString(status))
	}
	return uint32(resp.PathType), nil
}
