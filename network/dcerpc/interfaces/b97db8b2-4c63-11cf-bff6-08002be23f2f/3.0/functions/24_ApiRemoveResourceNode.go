package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiRemoveResourceNodeRequest carries the [in] parameters of ApiRemoveResourceNode.
type apiRemoveResourceNodeRequest struct {
	HResource mscmrp.HRES_RPC
	HNode     mscmrp.HNODE_RPC
}

func (*apiRemoveResourceNodeRequest) Opnum() uint16 { return clusapi.OpnumApiRemoveResourceNode }

// apiRemoveResourceNodeResponse carries the [out] parameters and return value of ApiRemoveResourceNode.
type apiRemoveResourceNodeResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiRemoveResourceNode calls ApiRemoveResourceNode (opnum 24) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiRemoveResourceNode(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, hNode mscmrp.HNODE_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiRemoveResourceNodeRequest{
		HResource: hResource,
		HNode:     hNode,
	}
	var resp apiRemoveResourceNodeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiRemoveResourceNode: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiRemoveResourceNode failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
