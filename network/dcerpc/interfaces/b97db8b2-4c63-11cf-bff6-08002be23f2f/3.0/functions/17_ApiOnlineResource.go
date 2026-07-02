package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOnlineResourceRequest carries the [in] parameters of ApiOnlineResource.
type apiOnlineResourceRequest struct {
	HResource mscmrp.HRES_RPC
}

func (*apiOnlineResourceRequest) Opnum() uint16 { return clusapi.OpnumApiOnlineResource }

// apiOnlineResourceResponse carries the [out] parameters and return value of ApiOnlineResource.
type apiOnlineResourceResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiOnlineResource calls ApiOnlineResource (opnum 17) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOnlineResource(rpc ndr.Invoker, hResource mscmrp.HRES_RPC) (Rpc_status ndr.DWORD, err error) {
	req := &apiOnlineResourceRequest{
		HResource: hResource,
	}
	var resp apiOnlineResourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOnlineResource: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOnlineResource failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
