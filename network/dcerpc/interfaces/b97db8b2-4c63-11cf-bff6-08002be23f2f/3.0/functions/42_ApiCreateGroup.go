package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateGroupRequest carries the [in] parameters of ApiCreateGroup.
type apiCreateGroupRequest struct {
	LpszGroupName ndr.WSTR
}

func (*apiCreateGroupRequest) Opnum() uint16 { return clusapi.OpnumApiCreateGroup }

// apiCreateGroupResponse carries the [out] parameters and return value of ApiCreateGroup.
type apiCreateGroupResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HGROUP_RPC `ndr:"retval"`
}

// ApiCreateGroup calls ApiCreateGroup (opnum 42) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateGroup(rpc ndr.Invoker, lpszGroupName ndr.WSTR) (Handle mscmrp.HGROUP_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateGroupRequest{
		LpszGroupName: lpszGroupName,
	}
	var resp apiCreateGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateGroup: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
