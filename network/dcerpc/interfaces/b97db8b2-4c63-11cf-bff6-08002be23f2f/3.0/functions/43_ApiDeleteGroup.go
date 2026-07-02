package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiDeleteGroupRequest carries the [in] parameters of ApiDeleteGroup.
type apiDeleteGroupRequest struct {
	Group mscmrp.HGROUP_RPC
	Force ndr.BOOL
}

func (*apiDeleteGroupRequest) Opnum() uint16 { return clusapi.OpnumApiDeleteGroup }

// apiDeleteGroupResponse carries the [out] parameters and return value of ApiDeleteGroup.
type apiDeleteGroupResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiDeleteGroup calls ApiDeleteGroup (opnum 43) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiDeleteGroup(rpc ndr.Invoker, group mscmrp.HGROUP_RPC, force ndr.BOOL) (Rpc_status ndr.DWORD, err error) {
	req := &apiDeleteGroupRequest{
		Group: group,
		Force: force,
	}
	var resp apiDeleteGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiDeleteGroup: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiDeleteGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
