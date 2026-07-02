package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenNodeRequest carries the [in] parameters of ApiOpenNode.
type apiOpenNodeRequest struct {
	LpszNodeName ndr.WSTR
}

func (*apiOpenNodeRequest) Opnum() uint16 { return clusapi.OpnumApiOpenNode }

// apiOpenNodeResponse carries the [out] parameters and return value of ApiOpenNode.
type apiOpenNodeResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HNODE_RPC `ndr:"retval"`
}

// ApiOpenNode calls ApiOpenNode (opnum 66) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenNode(rpc ndr.Invoker, lpszNodeName ndr.WSTR) (Handle mscmrp.HNODE_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenNodeRequest{
		LpszNodeName: lpszNodeName,
	}
	var resp apiOpenNodeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenNode: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenNode failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
