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

// apiChangeCsvStateExRequest carries the [in] parameters of ApiChangeCsvStateEx.
type apiChangeCsvStateExRequest struct {
	HResource      mscmrp.HRES_RPC
	DwState        ndr.DWORD
	LpszVolumeName ndr.WSTR
}

func (*apiChangeCsvStateExRequest) Opnum() uint16 { return clusapi.OpnumApiChangeCsvStateEx }

// apiChangeCsvStateExResponse carries the [out] parameters and return value of ApiChangeCsvStateEx.
type apiChangeCsvStateExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiChangeCsvStateEx calls ApiChangeCsvStateEx (opnum 182) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiChangeCsvStateEx(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, dwState ndr.DWORD, lpszVolumeName ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiChangeCsvStateExRequest{
		HResource:      hResource,
		DwState:        dwState,
		LpszVolumeName: lpszVolumeName,
	}
	var resp apiChangeCsvStateExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiChangeCsvStateEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiChangeCsvStateEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
