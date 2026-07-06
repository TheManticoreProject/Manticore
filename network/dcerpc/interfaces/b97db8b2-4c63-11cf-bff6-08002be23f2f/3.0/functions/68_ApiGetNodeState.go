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

// apiGetNodeStateRequest carries the [in] parameters of ApiGetNodeState.
type apiGetNodeStateRequest struct {
	HNode mscmrp.HNODE_RPC
}

func (*apiGetNodeStateRequest) Opnum() uint16 { return clusapi.OpnumApiGetNodeState }

// apiGetNodeStateResponse carries the [out] parameters and return value of ApiGetNodeState.
type apiGetNodeStateResponse struct {
	State      ndr.DWORD
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetNodeState calls ApiGetNodeState (opnum 68) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNodeState(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC) (State ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiGetNodeStateRequest{
		HNode: hNode,
	}
	var resp apiGetNodeStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNodeState: %w", err)
		return
	}
	State = resp.State
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNodeState failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
