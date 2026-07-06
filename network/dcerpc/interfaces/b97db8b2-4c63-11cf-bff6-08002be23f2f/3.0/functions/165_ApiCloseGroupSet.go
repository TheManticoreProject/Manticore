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

// apiCloseGroupSetRequest carries the [in] parameters of ApiCloseGroupSet.
type apiCloseGroupSetRequest struct {
	GroupSet mscmrp.HGROUPSET_RPC
}

func (*apiCloseGroupSetRequest) Opnum() uint16 { return clusapi.OpnumApiCloseGroupSet }

// apiCloseGroupSetResponse carries the [out] parameters and return value of ApiCloseGroupSet.
type apiCloseGroupSetResponse struct {
	GroupSet mscmrp.HGROUPSET_RPC
	Status   ndr.DWORD `ndr:"retval"`
}

// ApiCloseGroupSet calls ApiCloseGroupSet (opnum 165) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseGroupSet(rpc ndr.Invoker, groupSet mscmrp.HGROUPSET_RPC) (GroupSet mscmrp.HGROUPSET_RPC, err error) {
	req := &apiCloseGroupSetRequest{
		GroupSet: groupSet,
	}
	var resp apiCloseGroupSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseGroupSet: %w", err)
		return
	}
	GroupSet = resp.GroupSet
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseGroupSet failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
