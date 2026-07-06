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

// apiChangeResourceGroupExRequest carries the [in] parameters of ApiChangeResourceGroupEx.
type apiChangeResourceGroupExRequest struct {
	HResource mscmrp.HRES_RPC
	HGroup    mscmrp.HGROUP_RPC
	Flags     uint64
}

func (*apiChangeResourceGroupExRequest) Opnum() uint16 { return clusapi.OpnumApiChangeResourceGroupEx }

// apiChangeResourceGroupExResponse carries the [out] parameters and return value of ApiChangeResourceGroupEx.
type apiChangeResourceGroupExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiChangeResourceGroupEx calls ApiChangeResourceGroupEx (opnum 184) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiChangeResourceGroupEx(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, hGroup mscmrp.HGROUP_RPC, flags uint64) (Rpc_status ndr.DWORD, err error) {
	req := &apiChangeResourceGroupExRequest{
		HResource: hResource,
		HGroup:    hGroup,
		Flags:     flags,
	}
	var resp apiChangeResourceGroupExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiChangeResourceGroupEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiChangeResourceGroupEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
