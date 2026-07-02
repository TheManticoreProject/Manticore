package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddNotifyResourceRequest carries the [in] parameters of ApiAddNotifyResource.
type apiAddNotifyResourceRequest struct {
	HNotify     mscmrp.HNOTIFY_RPC
	HResource   mscmrp.HRES_RPC
	DwFilter    ndr.DWORD
	DwNotifyKey ndr.DWORD
}

func (*apiAddNotifyResourceRequest) Opnum() uint16 { return clusapi.OpnumApiAddNotifyResource }

// apiAddNotifyResourceResponse carries the [out] parameters and return value of ApiAddNotifyResource.
type apiAddNotifyResourceResponse struct {
	DwStateSequence ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiAddNotifyResource calls ApiAddNotifyResource (opnum 60) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddNotifyResource(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hResource mscmrp.HRES_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD) (DwStateSequence ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiAddNotifyResourceRequest{
		HNotify:     hNotify,
		HResource:   hResource,
		DwFilter:    dwFilter,
		DwNotifyKey: dwNotifyKey,
	}
	var resp apiAddNotifyResourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddNotifyResource: %w", err)
		return
	}
	DwStateSequence = resp.DwStateSequence
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddNotifyResource failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
