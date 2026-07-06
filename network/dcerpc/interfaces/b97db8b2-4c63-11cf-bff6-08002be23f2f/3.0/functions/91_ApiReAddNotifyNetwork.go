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

// apiReAddNotifyNetworkRequest carries the [in] parameters of ApiReAddNotifyNetwork.
type apiReAddNotifyNetworkRequest struct {
	HNotify       mscmrp.HNOTIFY_RPC
	HNetwork      mscmrp.HNETWORK_RPC
	DwFilter      ndr.DWORD
	DwNotifyKey   ndr.DWORD
	StateSequence ndr.DWORD
}

func (*apiReAddNotifyNetworkRequest) Opnum() uint16 { return clusapi.OpnumApiReAddNotifyNetwork }

// apiReAddNotifyNetworkResponse carries the [out] parameters and return value of ApiReAddNotifyNetwork.
type apiReAddNotifyNetworkResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiReAddNotifyNetwork calls ApiReAddNotifyNetwork (opnum 91) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiReAddNotifyNetwork(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hNetwork mscmrp.HNETWORK_RPC, dwFilter ndr.DWORD, dwNotifyKey ndr.DWORD, stateSequence ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiReAddNotifyNetworkRequest{
		HNotify:       hNotify,
		HNetwork:      hNetwork,
		DwFilter:      dwFilter,
		DwNotifyKey:   dwNotifyKey,
		StateSequence: stateSequence,
	}
	var resp apiReAddNotifyNetworkResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiReAddNotifyNetwork: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiReAddNotifyNetwork failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
