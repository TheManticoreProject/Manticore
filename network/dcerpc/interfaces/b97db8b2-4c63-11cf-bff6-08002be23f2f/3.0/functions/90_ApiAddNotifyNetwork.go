package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddNotifyNetworkRequest carries the [in] parameters of ApiAddNotifyNetwork.
type apiAddNotifyNetworkRequest struct {
	HNotify     mscmrp.HNOTIFY_RPC
	HNetwork    mscmrp.HNETWORK_RPC
	DwFilter    ndr.DWORD
	DwNotifyKey ndr.DWORD
}

func (*apiAddNotifyNetworkRequest) Opnum() uint16 { return clusapi.OpnumApiAddNotifyNetwork }

// apiAddNotifyNetworkResponse carries the [out] parameters and return value of ApiAddNotifyNetwork.
type apiAddNotifyNetworkResponse struct {
	DwStateSequence ndr.DWORD
	Rpc_status      ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiAddNotifyNetwork calls ApiAddNotifyNetwork (opnum 90) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddNotifyNetwork(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hNetwork mscmrp.HNETWORK_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD) (DwStateSequence ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiAddNotifyNetworkRequest{
		HNotify:     hNotify,
		HNetwork:    hNetwork,
		DwFilter:    dwFilter,
		DwNotifyKey: dwNotifyKey,
	}
	var resp apiAddNotifyNetworkResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddNotifyNetwork: %w", err)
		return
	}
	DwStateSequence = resp.DwStateSequence
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddNotifyNetwork failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
