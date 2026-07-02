package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiPauseNodeRequest carries the [in] parameters of ApiPauseNode.
type apiPauseNodeRequest struct {
	HNode mscmrp.HNODE_RPC
}

func (*apiPauseNodeRequest) Opnum() uint16 { return clusapi.OpnumApiPauseNode }

// apiPauseNodeResponse carries the [out] parameters and return value of ApiPauseNode.
type apiPauseNodeResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiPauseNode calls ApiPauseNode (opnum 69) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiPauseNode(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiPauseNodeRequest{
		HNode: hNode,
	}
	var resp apiPauseNodeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiPauseNode: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiPauseNode failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
