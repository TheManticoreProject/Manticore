package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCancelClusterGroupOperationRequest carries the [in] parameters of ApiCancelClusterGroupOperation.
type apiCancelClusterGroupOperationRequest struct {
	HGroup        mscmrp.HGROUP_RPC
	DwCancelFlags ndr.DWORD
}

func (*apiCancelClusterGroupOperationRequest) Opnum() uint16 {
	return clusapi.OpnumApiCancelClusterGroupOperation
}

// apiCancelClusterGroupOperationResponse carries the [out] parameters and return value of ApiCancelClusterGroupOperation.
type apiCancelClusterGroupOperationResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCancelClusterGroupOperation calls ApiCancelClusterGroupOperation (opnum 134) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCancelClusterGroupOperation(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, dwCancelFlags ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiCancelClusterGroupOperationRequest{
		HGroup:        hGroup,
		DwCancelFlags: dwCancelFlags,
	}
	var resp apiCancelClusterGroupOperationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCancelClusterGroupOperation: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCancelClusterGroupOperation failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
