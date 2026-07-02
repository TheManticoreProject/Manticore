package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiRemoveClusterGroupToGroupSetDependencyRequest carries the [in] parameters of ApiRemoveClusterGroupToGroupSetDependency.
type apiRemoveClusterGroupToGroupSetDependencyRequest struct {
	HGroup     mscmrp.HGROUP_RPC
	HDependsOn mscmrp.HGROUPSET_RPC
}

func (*apiRemoveClusterGroupToGroupSetDependencyRequest) Opnum() uint16 {
	return clusapi.OpnumApiRemoveClusterGroupToGroupSetDependency
}

// apiRemoveClusterGroupToGroupSetDependencyResponse carries the [out] parameters and return value of ApiRemoveClusterGroupToGroupSetDependency.
type apiRemoveClusterGroupToGroupSetDependencyResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiRemoveClusterGroupToGroupSetDependency calls ApiRemoveClusterGroupToGroupSetDependency (opnum 179) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiRemoveClusterGroupToGroupSetDependency(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, hDependsOn mscmrp.HGROUPSET_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiRemoveClusterGroupToGroupSetDependencyRequest{
		HGroup:     hGroup,
		HDependsOn: hDependsOn,
	}
	var resp apiRemoveClusterGroupToGroupSetDependencyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiRemoveClusterGroupToGroupSetDependency: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiRemoveClusterGroupToGroupSetDependency failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
