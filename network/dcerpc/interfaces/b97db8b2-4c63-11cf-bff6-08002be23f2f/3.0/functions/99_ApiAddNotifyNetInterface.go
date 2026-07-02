package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddNotifyNetInterfaceRequest carries the [in] parameters of ApiAddNotifyNetInterface.
type apiAddNotifyNetInterfaceRequest struct {
	HNotify       mscmrp.HNOTIFY_RPC
	HNetInterface mscmrp.HNETINTERFACE_RPC
	DwFilter      ndr.DWORD
	DwNotifyKey   ndr.DWORD
}

func (*apiAddNotifyNetInterfaceRequest) Opnum() uint16 { return clusapi.OpnumApiAddNotifyNetInterface }

// apiAddNotifyNetInterfaceResponse carries the [out] parameters and return value of ApiAddNotifyNetInterface.
type apiAddNotifyNetInterfaceResponse struct {
	DwStateSequence ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiAddNotifyNetInterface calls ApiAddNotifyNetInterface (opnum 99) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddNotifyNetInterface(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hNetInterface mscmrp.HNETINTERFACE_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD) (DwStateSequence ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiAddNotifyNetInterfaceRequest{
		HNotify:       hNotify,
		HNetInterface: hNetInterface,
		DwFilter:      dwFilter,
		DwNotifyKey:   dwNotifyKey,
	}
	var resp apiAddNotifyNetInterfaceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddNotifyNetInterface: %w", err)
		return
	}
	DwStateSequence = resp.DwStateSequence
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddNotifyNetInterface failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
