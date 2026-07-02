package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiMoveGroupRequest carries the [in] parameters of ApiMoveGroup.
type apiMoveGroupRequest struct {
	HGroup mscmrp.HGROUP_RPC
}

func (*apiMoveGroupRequest) Opnum() uint16 { return clusapi.OpnumApiMoveGroup }

// apiMoveGroupResponse carries the [out] parameters and return value of ApiMoveGroup.
type apiMoveGroupResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiMoveGroup calls ApiMoveGroup (opnum 51) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiMoveGroup(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiMoveGroupRequest{
		HGroup: hGroup,
	}
	var resp apiMoveGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiMoveGroup: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiMoveGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
