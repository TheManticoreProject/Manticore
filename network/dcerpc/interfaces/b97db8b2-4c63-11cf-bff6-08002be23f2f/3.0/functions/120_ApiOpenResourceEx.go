package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenResourceExRequest carries the [in] parameters of ApiOpenResourceEx.
type apiOpenResourceExRequest struct {
	LpszResourceName ndr.WSTR
	DwDesiredAccess  ndr.DWORD
}

func (*apiOpenResourceExRequest) Opnum() uint16 { return clusapi.OpnumApiOpenResourceEx }

// apiOpenResourceExResponse carries the [out] parameters and return value of ApiOpenResourceEx.
type apiOpenResourceExResponse struct {
	LpdwGrantedAccess ndr.DWORD
	Status            ndr.DWORD
	Rpc_status        ndr.DWORD
	Handle            mscmrp.HRES_RPC `ndr:"retval"`
}

// ApiOpenResourceEx calls ApiOpenResourceEx (opnum 120) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenResourceEx(rpc ndr.Invoker, lpszResourceName ndr.WSTR, dwDesiredAccess ndr.DWORD) (Handle mscmrp.HRES_RPC, LpdwGrantedAccess ndr.DWORD, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenResourceExRequest{
		LpszResourceName: lpszResourceName,
		DwDesiredAccess:  dwDesiredAccess,
	}
	var resp apiOpenResourceExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenResourceEx: %w", err)
		return
	}
	Handle = resp.Handle
	LpdwGrantedAccess = resp.LpdwGrantedAccess
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenResourceEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
