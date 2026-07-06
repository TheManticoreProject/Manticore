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

// apiChangeCsvStateRequest carries the [in] parameters of ApiChangeCsvState.
type apiChangeCsvStateRequest struct {
	HResource mscmrp.HRES_RPC
	DwState   ndr.DWORD
}

func (*apiChangeCsvStateRequest) Opnum() uint16 { return clusapi.OpnumApiChangeCsvState }

// apiChangeCsvStateResponse carries the [out] parameters and return value of ApiChangeCsvState.
type apiChangeCsvStateResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiChangeCsvState calls ApiChangeCsvState (opnum 123) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiChangeCsvState(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, dwState ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiChangeCsvStateRequest{
		HResource: hResource,
		DwState:   dwState,
	}
	var resp apiChangeCsvStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiChangeCsvState: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiChangeCsvState failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
