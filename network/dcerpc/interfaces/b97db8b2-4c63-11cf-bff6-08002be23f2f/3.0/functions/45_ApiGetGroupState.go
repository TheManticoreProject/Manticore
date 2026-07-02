package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetGroupStateRequest carries the [in] parameters of ApiGetGroupState.
type apiGetGroupStateRequest struct {
	HGroup mscmrp.HGROUP_RPC
}

func (*apiGetGroupStateRequest) Opnum() uint16 { return clusapi.OpnumApiGetGroupState }

// apiGetGroupStateResponse carries the [out] parameters and return value of ApiGetGroupState.
type apiGetGroupStateResponse struct {
	State      ndr.DWORD
	NodeName   ndr.WSTR
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetGroupState calls ApiGetGroupState (opnum 45) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetGroupState(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC) (State ndr.DWORD, NodeName ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetGroupStateRequest{
		HGroup: hGroup,
	}
	var resp apiGetGroupStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetGroupState: %w", err)
		return
	}
	State = resp.State
	NodeName = resp.NodeName
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetGroupState failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
