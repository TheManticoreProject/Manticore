package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiAddGroupSetDependencyRequest carries the [in] parameters of ApiAddGroupSetDependency.
type apiAddGroupSetDependencyRequest struct {
	DependentGroupSet mscmrp.HGROUPSET_RPC
	ProviderGroupSet  mscmrp.HGROUPSET_RPC
}

func (*apiAddGroupSetDependencyRequest) Opnum() uint16 { return clusapi.OpnumApiAddGroupSetDependency }

// apiAddGroupSetDependencyResponse carries the [out] parameters and return value of ApiAddGroupSetDependency.
type apiAddGroupSetDependencyResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiAddGroupSetDependency calls ApiAddGroupSetDependency (opnum 171) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddGroupSetDependency(rpc ndr.Invoker, dependentGroupSet mscmrp.HGROUPSET_RPC, providerGroupSet mscmrp.HGROUPSET_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiAddGroupSetDependencyRequest{
		DependentGroupSet: dependentGroupSet,
		ProviderGroupSet:  providerGroupSet,
	}
	var resp apiAddGroupSetDependencyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddGroupSetDependency: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddGroupSetDependency failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
