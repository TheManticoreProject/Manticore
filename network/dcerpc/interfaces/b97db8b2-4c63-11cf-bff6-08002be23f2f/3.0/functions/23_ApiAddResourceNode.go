package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddResourceNodeRequest carries the [in] parameters of ApiAddResourceNode.
type apiAddResourceNodeRequest struct {
	HResource mscmrp.HRES_RPC
	HNode     mscmrp.HNODE_RPC
}

func (*apiAddResourceNodeRequest) Opnum() uint16 { return clusapi.OpnumApiAddResourceNode }

// apiAddResourceNodeResponse carries the [out] parameters and return value of ApiAddResourceNode.
type apiAddResourceNodeResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiAddResourceNode calls ApiAddResourceNode (opnum 23) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddResourceNode(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, hNode mscmrp.HNODE_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiAddResourceNodeRequest{
		HResource: hResource,
		HNode:     hNode,
	}
	var resp apiAddResourceNodeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddResourceNode: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddResourceNode failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
