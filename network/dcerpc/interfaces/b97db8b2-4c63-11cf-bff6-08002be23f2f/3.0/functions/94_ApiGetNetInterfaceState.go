package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetNetInterfaceStateRequest carries the [in] parameters of ApiGetNetInterfaceState.
type apiGetNetInterfaceStateRequest struct {
	HNetInterface mscmrp.HNETINTERFACE_RPC
}

func (*apiGetNetInterfaceStateRequest) Opnum() uint16 { return clusapi.OpnumApiGetNetInterfaceState }

// apiGetNetInterfaceStateResponse carries the [out] parameters and return value of ApiGetNetInterfaceState.
type apiGetNetInterfaceStateResponse struct {
	State      ndr.DWORD
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetNetInterfaceState calls ApiGetNetInterfaceState (opnum 94) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNetInterfaceState(rpc ndr.Invoker, hNetInterface mscmrp.HNETINTERFACE_RPC) (State ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiGetNetInterfaceStateRequest{
		HNetInterface: hNetInterface,
	}
	var resp apiGetNetInterfaceStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNetInterfaceState: %w", err)
		return
	}
	State = resp.State
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNetInterfaceState failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
