package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiRemoveClusterGroupDependencyRequest carries the [in] parameters of ApiRemoveClusterGroupDependency.
type apiRemoveClusterGroupDependencyRequest struct {
	HGroup     mscmrp.HGROUP_RPC
	HDependsOn mscmrp.HGROUP_RPC
}

func (*apiRemoveClusterGroupDependencyRequest) Opnum() uint16 {
	return clusapi.OpnumApiRemoveClusterGroupDependency
}

// apiRemoveClusterGroupDependencyResponse carries the [out] parameters and return value of ApiRemoveClusterGroupDependency.
type apiRemoveClusterGroupDependencyResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiRemoveClusterGroupDependency calls ApiRemoveClusterGroupDependency (opnum 176) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiRemoveClusterGroupDependency(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, hDependsOn mscmrp.HGROUP_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiRemoveClusterGroupDependencyRequest{
		HGroup:     hGroup,
		HDependsOn: hDependsOn,
	}
	var resp apiRemoveClusterGroupDependencyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiRemoveClusterGroupDependency: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiRemoveClusterGroupDependency failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
