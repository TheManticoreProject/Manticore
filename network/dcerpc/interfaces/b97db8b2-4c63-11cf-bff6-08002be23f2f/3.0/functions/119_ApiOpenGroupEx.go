package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenGroupExRequest carries the [in] parameters of ApiOpenGroupEx.
type apiOpenGroupExRequest struct {
	LpszGroupName   ndr.WSTR
	DwDesiredAccess ndr.DWORD
}

func (*apiOpenGroupExRequest) Opnum() uint16 { return clusapi.OpnumApiOpenGroupEx }

// apiOpenGroupExResponse carries the [out] parameters and return value of ApiOpenGroupEx.
type apiOpenGroupExResponse struct {
	LpdwGrantedAccess ndr.DWORD
	Status            ndr.DWORD
	Rpc_status        ndr.DWORD
	Handle            mscmrp.HGROUP_RPC `ndr:"retval"`
}

// ApiOpenGroupEx calls ApiOpenGroupEx (opnum 119) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenGroupEx(rpc ndr.Invoker, lpszGroupName ndr.WSTR, dwDesiredAccess ndr.DWORD) (Handle mscmrp.HGROUP_RPC, LpdwGrantedAccess ndr.DWORD, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenGroupExRequest{
		LpszGroupName:   lpszGroupName,
		DwDesiredAccess: dwDesiredAccess,
	}
	var resp apiOpenGroupExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenGroupEx: %w", err)
		return
	}
	Handle = resp.Handle
	LpdwGrantedAccess = resp.LpdwGrantedAccess
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenGroupEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
