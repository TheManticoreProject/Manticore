package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiSetGroupNameRequest carries the [in] parameters of ApiSetGroupName.
type apiSetGroupNameRequest struct {
	HGroup        mscmrp.HGROUP_RPC
	LpszGroupName ndr.WSTR
}

func (*apiSetGroupNameRequest) Opnum() uint16 { return clusapi.OpnumApiSetGroupName }

// apiSetGroupNameResponse carries the [out] parameters and return value of ApiSetGroupName.
type apiSetGroupNameResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetGroupName calls ApiSetGroupName (opnum 46) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetGroupName(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, lpszGroupName ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetGroupNameRequest{
		HGroup:        hGroup,
		LpszGroupName: lpszGroupName,
	}
	var resp apiSetGroupNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetGroupName: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetGroupName failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
