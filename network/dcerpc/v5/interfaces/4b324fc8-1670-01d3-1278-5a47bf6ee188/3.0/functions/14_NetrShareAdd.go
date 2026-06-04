package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
)

// netrShareAddRequest is the [in] parameter set of NetrShareAdd: the optional server name,
// the info level, the inline [in, switch_is(Level)] SHARE_INFO union (its Tag is set to
// Level before marshalling), and the optional [in,out,unique] ParmErr pointer.
type netrShareAddRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	InfoStruct structures.SHARE_INFO
	ParmErr    *ndr.DWORD `ndr:"unique"`
}

func (*netrShareAddRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareAdd
}

// netrShareAddResponse is the reply: the [in,out,unique] ParmErr pointer and the
// NET_API_STATUS return value.
type netrShareAddResponse struct {
	ParmErr *ndr.DWORD `ndr:"unique"`
	Status  ndr.DWORD  `ndr:"retval"`
}

// NetrShareAdd calls NetrShareAdd (opnum 14), sharing a resource on the server
// ([MS-SRVS] 3.1.4.7). The union discriminant is set to level before marshalling.
func NetrShareAdd(rpc *client.Client, serverName string, level ndr.DWORD, infoStruct structures.SHARE_INFO, parmErr *ndr.DWORD) (*ndr.DWORD, error) {
	infoStruct.Tag = level
	req := &netrShareAddRequest{
		ServerName: optWStr(serverName),
		Level:      level,
		InfoStruct: infoStruct,
		ParmErr:    parmErr,
	}
	var resp netrShareAddResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("NetrShareAdd: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return resp.ParmErr, fmt.Errorf("NetrShareAdd failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return resp.ParmErr, nil
}
