package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenGroupRequest carries the [in] parameters of ApiOpenGroup.
type apiOpenGroupRequest struct {
	LpszGroupName ndr.WSTR
}

func (*apiOpenGroupRequest) Opnum() uint16 { return clusapi.OpnumApiOpenGroup }

// apiOpenGroupResponse carries the [out] parameters and return value of ApiOpenGroup.
type apiOpenGroupResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HGROUP_RPC `ndr:"retval"`
}

// ApiOpenGroup calls ApiOpenGroup (opnum 41) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenGroup(rpc ndr.Invoker, lpszGroupName ndr.WSTR) (Handle mscmrp.HGROUP_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenGroupRequest{
		LpszGroupName: lpszGroupName,
	}
	var resp apiOpenGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenGroup: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
