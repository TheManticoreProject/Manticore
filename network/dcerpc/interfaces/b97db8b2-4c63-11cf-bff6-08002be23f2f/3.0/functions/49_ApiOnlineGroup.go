package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOnlineGroupRequest carries the [in] parameters of ApiOnlineGroup.
type apiOnlineGroupRequest struct {
	HGroup mscmrp.HGROUP_RPC
}

func (*apiOnlineGroupRequest) Opnum() uint16 { return clusapi.OpnumApiOnlineGroup }

// apiOnlineGroupResponse carries the [out] parameters and return value of ApiOnlineGroup.
type apiOnlineGroupResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiOnlineGroup calls ApiOnlineGroup (opnum 49) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOnlineGroup(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiOnlineGroupRequest{
		HGroup: hGroup,
	}
	var resp apiOnlineGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOnlineGroup: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOnlineGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
