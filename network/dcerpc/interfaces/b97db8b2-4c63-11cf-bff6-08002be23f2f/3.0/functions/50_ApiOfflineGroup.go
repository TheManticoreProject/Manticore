package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOfflineGroupRequest carries the [in] parameters of ApiOfflineGroup.
type apiOfflineGroupRequest struct {
	HGroup mscmrp.HGROUP_RPC
}

func (*apiOfflineGroupRequest) Opnum() uint16 { return clusapi.OpnumApiOfflineGroup }

// apiOfflineGroupResponse carries the [out] parameters and return value of ApiOfflineGroup.
type apiOfflineGroupResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiOfflineGroup calls ApiOfflineGroup (opnum 50) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOfflineGroup(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiOfflineGroupRequest{
		HGroup: hGroup,
	}
	var resp apiOfflineGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOfflineGroup: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOfflineGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
