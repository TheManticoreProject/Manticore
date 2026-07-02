package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiEvictNodeRequest carries the [in] parameters of ApiEvictNode.
type apiEvictNodeRequest struct {
	HNode mscmrp.HNODE_RPC
}

func (*apiEvictNodeRequest) Opnum() uint16 { return clusapi.OpnumApiEvictNode }

// apiEvictNodeResponse carries the [out] parameters and return value of ApiEvictNode.
type apiEvictNodeResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiEvictNode calls ApiEvictNode (opnum 71) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiEvictNode(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiEvictNodeRequest{
		HNode: hNode,
	}
	var resp apiEvictNodeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiEvictNode: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiEvictNode failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
