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

// apiOpenResourceRequest carries the [in] parameters of ApiOpenResource.
type apiOpenResourceRequest struct {
	LpszResourceName ndr.WSTR
}

func (*apiOpenResourceRequest) Opnum() uint16 { return clusapi.OpnumApiOpenResource }

// apiOpenResourceResponse carries the [out] parameters and return value of ApiOpenResource.
type apiOpenResourceResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HRES_RPC `ndr:"retval"`
}

// ApiOpenResource calls ApiOpenResource (opnum 8) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenResource(rpc ndr.Invoker, lpszResourceName ndr.WSTR) (Handle mscmrp.HRES_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenResourceRequest{
		LpszResourceName: lpszResourceName,
	}
	var resp apiOpenResourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenResource: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenResource failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
