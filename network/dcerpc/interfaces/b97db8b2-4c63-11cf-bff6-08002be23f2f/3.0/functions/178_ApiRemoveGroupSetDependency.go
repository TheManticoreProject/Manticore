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

// apiRemoveGroupSetDependencyRequest carries the [in] parameters of ApiRemoveGroupSetDependency.
type apiRemoveGroupSetDependencyRequest struct {
	HGroupSet  mscmrp.HGROUPSET_RPC
	HDependsOn mscmrp.HGROUPSET_RPC
}

func (*apiRemoveGroupSetDependencyRequest) Opnum() uint16 {
	return clusapi.OpnumApiRemoveGroupSetDependency
}

// apiRemoveGroupSetDependencyResponse carries the [out] parameters and return value of ApiRemoveGroupSetDependency.
type apiRemoveGroupSetDependencyResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiRemoveGroupSetDependency calls ApiRemoveGroupSetDependency (opnum 178) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiRemoveGroupSetDependency(rpc ndr.Invoker, hGroupSet mscmrp.HGROUPSET_RPC, hDependsOn mscmrp.HGROUPSET_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiRemoveGroupSetDependencyRequest{
		HGroupSet:  hGroupSet,
		HDependsOn: hDependsOn,
	}
	var resp apiRemoveGroupSetDependencyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiRemoveGroupSetDependency: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiRemoveGroupSetDependency failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
