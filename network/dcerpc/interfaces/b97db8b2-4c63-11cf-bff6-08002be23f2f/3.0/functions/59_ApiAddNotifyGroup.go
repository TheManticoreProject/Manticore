package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddNotifyGroupRequest carries the [in] parameters of ApiAddNotifyGroup.
type apiAddNotifyGroupRequest struct {
	HNotify     mscmrp.HNOTIFY_RPC
	HGroup      mscmrp.HGROUP_RPC
	DwFilter    ndr.DWORD
	DwNotifyKey ndr.DWORD
}

func (*apiAddNotifyGroupRequest) Opnum() uint16 { return clusapi.OpnumApiAddNotifyGroup }

// apiAddNotifyGroupResponse carries the [out] parameters and return value of ApiAddNotifyGroup.
type apiAddNotifyGroupResponse struct {
	DwStateSequence ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiAddNotifyGroup calls ApiAddNotifyGroup (opnum 59) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddNotifyGroup(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hGroup mscmrp.HGROUP_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD) (DwStateSequence ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiAddNotifyGroupRequest{
		HNotify:     hNotify,
		HGroup:      hGroup,
		DwFilter:    dwFilter,
		DwNotifyKey: dwNotifyKey,
	}
	var resp apiAddNotifyGroupResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddNotifyGroup: %w", err)
		return
	}
	DwStateSequence = resp.DwStateSequence
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddNotifyGroup failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
