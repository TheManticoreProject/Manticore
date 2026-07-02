package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetKeySecurityRequest carries the [in] parameters of ApiGetKeySecurity.
type apiGetKeySecurityRequest struct {
	HKey                   mscmrp.HKEY_RPC
	SecurityInformation    ndr.DWORD
	PRpcSecurityDescriptor mscmrp.RPC_SECURITY_DESCRIPTOR
}

func (*apiGetKeySecurityRequest) Opnum() uint16 { return clusapi.OpnumApiGetKeySecurity }

// apiGetKeySecurityResponse carries the [out] parameters and return value of ApiGetKeySecurity.
type apiGetKeySecurityResponse struct {
	PRpcSecurityDescriptor mscmrp.RPC_SECURITY_DESCRIPTOR
	Rpc_status             ndr.DWORD
	Status                 ndr.DWORD `ndr:"retval"`
}

// ApiGetKeySecurity calls ApiGetKeySecurity (opnum 40) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetKeySecurity(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, securityInformation ndr.DWORD, pRpcSecurityDescriptor mscmrp.RPC_SECURITY_DESCRIPTOR) (PRpcSecurityDescriptor mscmrp.RPC_SECURITY_DESCRIPTOR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetKeySecurityRequest{
		HKey:                   hKey,
		SecurityInformation:    securityInformation,
		PRpcSecurityDescriptor: pRpcSecurityDescriptor,
	}
	var resp apiGetKeySecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetKeySecurity: %w", err)
		return
	}
	PRpcSecurityDescriptor = resp.PRpcSecurityDescriptor
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetKeySecurity failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
