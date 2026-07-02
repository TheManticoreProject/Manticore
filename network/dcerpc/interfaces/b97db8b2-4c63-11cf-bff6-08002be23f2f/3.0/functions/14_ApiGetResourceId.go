package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetResourceIdRequest carries the [in] parameters of ApiGetResourceId.
type apiGetResourceIdRequest struct {
	HResource mscmrp.HRES_RPC
}

func (*apiGetResourceIdRequest) Opnum() uint16 { return clusapi.OpnumApiGetResourceId }

// apiGetResourceIdResponse carries the [out] parameters and return value of ApiGetResourceId.
type apiGetResourceIdResponse struct {
	PGuid      ndr.WSTR
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetResourceId calls ApiGetResourceId (opnum 14) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetResourceId(rpc ndr.Invoker, hResource mscmrp.HRES_RPC) (PGuid ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetResourceIdRequest{
		HResource: hResource,
	}
	var resp apiGetResourceIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetResourceId: %w", err)
		return
	}
	PGuid = resp.PGuid
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetResourceId failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
