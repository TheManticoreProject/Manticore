package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetResourceStateRequest carries the [in] parameters of ApiGetResourceState.
type apiGetResourceStateRequest struct {
	HResource mscmrp.HRES_RPC
}

func (*apiGetResourceStateRequest) Opnum() uint16 { return clusapi.OpnumApiGetResourceState }

// apiGetResourceStateResponse carries the [out] parameters and return value of ApiGetResourceState.
type apiGetResourceStateResponse struct {
	State      ndr.DWORD
	NodeName   ndr.WSTR
	GroupName  ndr.WSTR
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetResourceState calls ApiGetResourceState (opnum 12) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetResourceState(rpc ndr.Invoker, hResource mscmrp.HRES_RPC) (State ndr.DWORD, NodeName ndr.WSTR, GroupName ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetResourceStateRequest{
		HResource: hResource,
	}
	var resp apiGetResourceStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetResourceState: %w", err)
		return
	}
	State = resp.State
	NodeName = resp.NodeName
	GroupName = resp.GroupName
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetResourceState failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
