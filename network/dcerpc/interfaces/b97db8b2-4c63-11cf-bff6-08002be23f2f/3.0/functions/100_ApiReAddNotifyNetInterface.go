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

// apiReAddNotifyNetInterfaceRequest carries the [in] parameters of ApiReAddNotifyNetInterface.
type apiReAddNotifyNetInterfaceRequest struct {
	HNotify       mscmrp.HNOTIFY_RPC
	HNetInterface mscmrp.HNETINTERFACE_RPC
	DwFilter      ndr.DWORD
	DwNotifyKey   ndr.DWORD
	StateSequence ndr.DWORD
}

func (*apiReAddNotifyNetInterfaceRequest) Opnum() uint16 {
	return clusapi.OpnumApiReAddNotifyNetInterface
}

// apiReAddNotifyNetInterfaceResponse carries the [out] parameters and return value of ApiReAddNotifyNetInterface.
type apiReAddNotifyNetInterfaceResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiReAddNotifyNetInterface calls ApiReAddNotifyNetInterface (opnum 100) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiReAddNotifyNetInterface(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hNetInterface mscmrp.HNETINTERFACE_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD, stateSequence ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiReAddNotifyNetInterfaceRequest{
		HNotify:       hNotify,
		HNetInterface: hNetInterface,
		DwFilter:      dwFilter,
		DwNotifyKey:   dwNotifyKey,
		StateSequence: stateSequence,
	}
	var resp apiReAddNotifyNetInterfaceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiReAddNotifyNetInterface: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiReAddNotifyNetInterface failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
