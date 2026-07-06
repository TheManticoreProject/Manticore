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

// apiReAddNotifyResourceRequest carries the [in] parameters of ApiReAddNotifyResource.
type apiReAddNotifyResourceRequest struct {
	HNotify       mscmrp.HNOTIFY_RPC
	HResource     mscmrp.HRES_RPC
	DwFilter      ndr.DWORD
	DwNotifyKey   ndr.DWORD
	StateSequence ndr.DWORD
}

func (*apiReAddNotifyResourceRequest) Opnum() uint16 { return clusapi.OpnumApiReAddNotifyResource }

// apiReAddNotifyResourceResponse carries the [out] parameters and return value of ApiReAddNotifyResource.
type apiReAddNotifyResourceResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiReAddNotifyResource calls ApiReAddNotifyResource (opnum 64) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiReAddNotifyResource(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hResource mscmrp.HRES_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD, stateSequence ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiReAddNotifyResourceRequest{
		HNotify:       hNotify,
		HResource:     hResource,
		DwFilter:      dwFilter,
		DwNotifyKey:   dwNotifyKey,
		StateSequence: stateSequence,
	}
	var resp apiReAddNotifyResourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiReAddNotifyResource: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiReAddNotifyResource failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
