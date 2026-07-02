package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddGroupToGroupSetDependencyRequest carries the [in] parameters of ApiAddGroupToGroupSetDependency.
type apiAddGroupToGroupSetDependencyRequest struct {
	DependentGroup   mscmrp.HGROUP_RPC
	ProviderGroupSet mscmrp.HGROUPSET_RPC
}

func (*apiAddGroupToGroupSetDependencyRequest) Opnum() uint16 {
	return clusapi.OpnumApiAddGroupToGroupSetDependency
}

// apiAddGroupToGroupSetDependencyResponse carries the [out] parameters and return value of ApiAddGroupToGroupSetDependency.
type apiAddGroupToGroupSetDependencyResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiAddGroupToGroupSetDependency calls ApiAddGroupToGroupSetDependency (opnum 172) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddGroupToGroupSetDependency(rpc ndr.Invoker, dependentGroup mscmrp.HGROUP_RPC, providerGroupSet mscmrp.HGROUPSET_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiAddGroupToGroupSetDependencyRequest{
		DependentGroup:   dependentGroup,
		ProviderGroupSet: providerGroupSet,
	}
	var resp apiAddGroupToGroupSetDependencyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddGroupToGroupSetDependency: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddGroupToGroupSetDependency failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
