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

// apiAddGroupToGroupSetRequest carries the [in] parameters of ApiAddGroupToGroupSet.
type apiAddGroupToGroupSetRequest struct {
	GroupSet mscmrp.HGROUPSET_RPC
	Group    mscmrp.HGROUP_RPC
}

func (*apiAddGroupToGroupSetRequest) Opnum() uint16 { return clusapi.OpnumApiAddGroupToGroupSet }

// apiAddGroupToGroupSetResponse carries the [out] parameters and return value of ApiAddGroupToGroupSet.
type apiAddGroupToGroupSetResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiAddGroupToGroupSet calls ApiAddGroupToGroupSet (opnum 167) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddGroupToGroupSet(rpc ndr.Invoker, groupSet mscmrp.HGROUPSET_RPC, group mscmrp.HGROUP_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiAddGroupToGroupSetRequest{
		GroupSet: groupSet,
		Group:    group,
	}
	var resp apiAddGroupToGroupSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddGroupToGroupSet: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddGroupToGroupSet failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
