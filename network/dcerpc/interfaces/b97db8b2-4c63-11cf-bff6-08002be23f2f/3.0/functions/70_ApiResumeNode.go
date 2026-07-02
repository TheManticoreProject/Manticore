package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiResumeNodeRequest carries the [in] parameters of ApiResumeNode.
type apiResumeNodeRequest struct {
	HNode mscmrp.HNODE_RPC
}

func (*apiResumeNodeRequest) Opnum() uint16 { return clusapi.OpnumApiResumeNode }

// apiResumeNodeResponse carries the [out] parameters and return value of ApiResumeNode.
type apiResumeNodeResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiResumeNode calls ApiResumeNode (opnum 70) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiResumeNode(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiResumeNodeRequest{
		HNode: hNode,
	}
	var resp apiResumeNodeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiResumeNode: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiResumeNode failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
