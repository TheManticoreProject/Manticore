package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
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
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCanResourceBeDependent failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
