package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCanResourceBeDependentRequest carries the [in] parameters of ApiCanResourceBeDependent.
type apiCanResourceBeDependentRequest struct {
	HResource          mscmrp.HRES_RPC
	HResourceDependent mscmrp.HRES_RPC
}

func (*apiCanResourceBeDependentRequest) Opnum() uint16 {
	return clusapi.OpnumApiCanResourceBeDependent
}

// apiCanResourceBeDependentResponse carries the [out] parameters and return value of ApiCanResourceBeDependent.
type apiCanResourceBeDependentResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCanResourceBeDependent calls ApiCanResourceBeDependent (opnum 21) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCanResourceBeDependent(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, hResourceDependent mscmrp.HRES_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiCanResourceBeDependentRequest{
		HResource:          hResource,
		HResourceDependent: hResourceDependent,
	}
	var resp apiCanResourceBeDependentResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCanResourceBeDependent: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("ApiCanResourceBeDependent failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
