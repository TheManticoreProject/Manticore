package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiSetServiceAccountPasswordRequest carries the [in] parameters of ApiSetServiceAccountPassword.
type apiSetServiceAccountPasswordRequest struct {
	LpszNewPassword        ndr.WSTR
	DwFlags                mscmrp.IDL_CLUSTER_SET_PASSWORD_FLAGS
	ReturnStatusBufferSize ndr.DWORD
}

func (*apiSetServiceAccountPasswordRequest) Opnum() uint16 {
	return clusapi.OpnumApiSetServiceAccountPassword
}

// apiSetServiceAccountPasswordResponse carries the [out] parameters and return value of ApiSetServiceAccountPassword.
type apiSetServiceAccountPasswordResponse struct {
	ReturnStatusBufferPtr []mscmrp.IDL_CLUSTER_SET_PASSWORD_STATUS `ndr:"ref,size_is=ReturnStatusBufferSize,varying"`
	SizeReturned          ndr.DWORD
	ExpectedBufferSize    ndr.DWORD
	Status                ndr.DWORD `ndr:"retval"`
}

// ApiSetServiceAccountPassword calls ApiSetServiceAccountPassword (opnum 108) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetServiceAccountPassword(rpc ndr.Invoker, lpszNewPassword ndr.WSTR, dwFlags mscmrp.IDL_CLUSTER_SET_PASSWORD_FLAGS, returnStatusBufferSize ndr.DWORD) (ReturnStatusBufferPtr []mscmrp.IDL_CLUSTER_SET_PASSWORD_STATUS, SizeReturned ndr.DWORD, ExpectedBufferSize ndr.DWORD, err error) {
	req := &apiSetServiceAccountPasswordRequest{
		LpszNewPassword:        lpszNewPassword,
		DwFlags:                dwFlags,
		ReturnStatusBufferSize: returnStatusBufferSize,
	}
	var resp apiSetServiceAccountPasswordResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetServiceAccountPassword: %w", err)
		return
	}
	ReturnStatusBufferPtr = resp.ReturnStatusBufferPtr
	SizeReturned = resp.SizeReturned
	ExpectedBufferSize = resp.ExpectedBufferSize
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetServiceAccountPassword failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
