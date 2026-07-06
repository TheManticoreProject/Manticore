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

// apiSetKeySecurityRequest carries the [in] parameters of ApiSetKeySecurity.
type apiSetKeySecurityRequest struct {
	HKey                   mscmrp.HKEY_RPC
	SecurityInformation    ndr.DWORD
	PRpcSecurityDescriptor mscmrp.RPC_SECURITY_DESCRIPTOR
}

func (*apiSetKeySecurityRequest) Opnum() uint16 { return clusapi.OpnumApiSetKeySecurity }

// apiSetKeySecurityResponse carries the [out] parameters and return value of ApiSetKeySecurity.
type apiSetKeySecurityResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetKeySecurity calls ApiSetKeySecurity (opnum 39) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetKeySecurity(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, securityInformation ndr.DWORD, pRpcSecurityDescriptor mscmrp.RPC_SECURITY_DESCRIPTOR) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetKeySecurityRequest{
		HKey:                   hKey,
		SecurityInformation:    securityInformation,
		PRpcSecurityDescriptor: pRpcSecurityDescriptor,
	}
	var resp apiSetKeySecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetKeySecurity: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetKeySecurity failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
