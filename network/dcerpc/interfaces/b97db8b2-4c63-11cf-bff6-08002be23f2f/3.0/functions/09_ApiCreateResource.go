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

// apiCreateResourceRequest carries the [in] parameters of ApiCreateResource.
type apiCreateResourceRequest struct {
	HGroup           mscmrp.HGROUP_RPC
	LpszResourceName ndr.WSTR
	LpszResourceType ndr.WSTR
	DwFlags          ndr.DWORD
}

func (*apiCreateResourceRequest) Opnum() uint16 { return clusapi.OpnumApiCreateResource }

// apiCreateResourceResponse carries the [out] parameters and return value of ApiCreateResource.
type apiCreateResourceResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HRES_RPC `ndr:"retval"`
}

// ApiCreateResource calls ApiCreateResource (opnum 9) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateResource(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, lpszResourceName ndr.WSTR, lpszResourceType ndr.WSTR, dwFlags ndr.DWORD) (Handle mscmrp.HRES_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateResourceRequest{
		HGroup:           hGroup,
		LpszResourceName: lpszResourceName,
		LpszResourceType: lpszResourceType,
		DwFlags:          dwFlags,
	}
	var resp apiCreateResourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateResource: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateResource failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
