package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiReAddNotifyGroupRequest carries the [in] parameters of ApiReAddNotifyGroup.
type apiReAddNotifyGroupRequest struct {
	HNotify       mscmrp.HNOTIFY_RPC
	HGroup        mscmrp.HGROUP_RPC
	DwFilter      ndr.DWORD
	DwNotifyKey   ndr.DWORD
	StateSequence ndr.DWORD
}

func (*apiReAddNotifyGroupRequest) Opnum() uint16 { return clusapi.OpnumApiReAddNotifyGroup }

// apiReAddNotifyGroupResponse carries the [out] parameters and return value of ApiReAddNotifyGroup.
type apiReAddNotifyGroupResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiReAddNotifyGroup calls ApiReAddNotifyGroup (opnum 63) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiReAddNotifyGroup(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hGroup mscmrp.HGROUP_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD, stateSequence ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiReAddNotifyGroupRequest{
		HNotify:       hNotify,
		HGroup:        hGroup,
		DwFilter:      dwFilter,
		DwNotifyKey:   dwNotifyKey,
		StateSequence: stateSequence,
	}
	var resp apiReAddNotifyGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiReAddNotifyGroup: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiReAddNotifyGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
