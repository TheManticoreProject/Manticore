package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateGroupSetRequest carries the [in] parameters of ApiCreateGroupSet.
type apiCreateGroupSetRequest struct {
	LpszGroupSetName ndr.WSTR
}

func (*apiCreateGroupSetRequest) Opnum() uint16 { return clusapi.OpnumApiCreateGroupSet }

// apiCreateGroupSetResponse carries the [out] parameters and return value of ApiCreateGroupSet.
type apiCreateGroupSetResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HGROUPSET_RPC `ndr:"retval"`
}

// ApiCreateGroupSet calls ApiCreateGroupSet (opnum 163) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateGroupSet(rpc ndr.Invoker, lpszGroupSetName ndr.WSTR) (Handle mscmrp.HGROUPSET_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateGroupSetRequest{
		LpszGroupSetName: lpszGroupSetName,
	}
	var resp apiCreateGroupSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateGroupSet: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateGroupSet failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
