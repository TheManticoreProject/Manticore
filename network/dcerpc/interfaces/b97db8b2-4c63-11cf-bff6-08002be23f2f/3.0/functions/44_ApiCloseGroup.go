package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCloseGroupRequest carries the [in] parameters of ApiCloseGroup.
type apiCloseGroupRequest struct {
	Group mscmrp.HGROUP_RPC
}

func (*apiCloseGroupRequest) Opnum() uint16 { return clusapi.OpnumApiCloseGroup }

// apiCloseGroupResponse carries the [out] parameters and return value of ApiCloseGroup.
type apiCloseGroupResponse struct {
	Group  mscmrp.HGROUP_RPC
	Status ndr.DWORD `ndr:"retval"`
}

// ApiCloseGroup calls ApiCloseGroup (opnum 44) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseGroup(rpc ndr.Invoker, group mscmrp.HGROUP_RPC) (Group mscmrp.HGROUP_RPC, err error) {
	req := &apiCloseGroupRequest{
		Group: group,
	}
	var resp apiCloseGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseGroup: %w", err)
		return
	}
	Group = resp.Group
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
