package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiDeleteGroupSetRequest carries the [in] parameters of ApiDeleteGroupSet.
type apiDeleteGroupSetRequest struct {
	GroupSet mscmrp.HGROUPSET_RPC
}

func (*apiDeleteGroupSetRequest) Opnum() uint16 { return clusapi.OpnumApiDeleteGroupSet }

// apiDeleteGroupSetResponse carries the [out] parameters and return value of ApiDeleteGroupSet.
type apiDeleteGroupSetResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiDeleteGroupSet calls ApiDeleteGroupSet (opnum 166) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiDeleteGroupSet(rpc ndr.Invoker, groupSet mscmrp.HGROUPSET_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiDeleteGroupSetRequest{
		GroupSet: groupSet,
	}
	var resp apiDeleteGroupSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiDeleteGroupSet: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiDeleteGroupSet failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
