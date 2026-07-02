package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetNetworkStateRequest carries the [in] parameters of ApiGetNetworkState.
type apiGetNetworkStateRequest struct {
	HNetwork mscmrp.HNETWORK_RPC
}

func (*apiGetNetworkStateRequest) Opnum() uint16 { return clusapi.OpnumApiGetNetworkState }

// apiGetNetworkStateResponse carries the [out] parameters and return value of ApiGetNetworkState.
type apiGetNetworkStateResponse struct {
	State      ndr.DWORD
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetNetworkState calls ApiGetNetworkState (opnum 83) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNetworkState(rpc ndr.Invoker, hNetwork mscmrp.HNETWORK_RPC) (State ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiGetNetworkStateRequest{
		HNetwork: hNetwork,
	}
	var resp apiGetNetworkStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNetworkState: %w", err)
		return
	}
	State = resp.State
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNetworkState failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
