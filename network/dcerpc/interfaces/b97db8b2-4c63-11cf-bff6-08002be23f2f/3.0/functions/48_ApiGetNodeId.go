package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetNodeIdRequest carries the [in] parameters of ApiGetNodeId.
type apiGetNodeIdRequest struct {
	HNode mscmrp.HNODE_RPC
}

func (*apiGetNodeIdRequest) Opnum() uint16 { return clusapi.OpnumApiGetNodeId }

// apiGetNodeIdResponse carries the [out] parameters and return value of ApiGetNodeId.
type apiGetNodeIdResponse struct {
	PGuid      ndr.WSTR
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetNodeId calls ApiGetNodeId (opnum 48) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNodeId(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC) (PGuid ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetNodeIdRequest{
		HNode: hNode,
	}
	var resp apiGetNodeIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNodeId: %w", err)
		return
	}
	PGuid = resp.PGuid
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNodeId failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
