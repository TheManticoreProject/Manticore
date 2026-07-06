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

// apiCloseResourceRequest carries the [in] parameters of ApiCloseResource.
type apiCloseResourceRequest struct {
	Resource mscmrp.HRES_RPC
}

func (*apiCloseResourceRequest) Opnum() uint16 { return clusapi.OpnumApiCloseResource }

// apiCloseResourceResponse carries the [out] parameters and return value of ApiCloseResource.
type apiCloseResourceResponse struct {
	Resource mscmrp.HRES_RPC
	Status   ndr.DWORD `ndr:"retval"`
}

// ApiCloseResource calls ApiCloseResource (opnum 11) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseResource(rpc ndr.Invoker, resource mscmrp.HRES_RPC) (Resource mscmrp.HRES_RPC, err error) {
	req := &apiCloseResourceRequest{
		Resource: resource,
	}
	var resp apiCloseResourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseResource: %w", err)
		return
	}
	Resource = resp.Resource
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseResource failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
