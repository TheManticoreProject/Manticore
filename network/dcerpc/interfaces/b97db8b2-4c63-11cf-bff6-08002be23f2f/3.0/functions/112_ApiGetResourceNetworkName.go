package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetResourceNetworkNameRequest carries the [in] parameters of ApiGetResourceNetworkName.
type apiGetResourceNetworkNameRequest struct {
	HResource mscmrp.HRES_RPC
}

func (*apiGetResourceNetworkNameRequest) Opnum() uint16 {
	return clusapi.OpnumApiGetResourceNetworkName
}

// apiGetResourceNetworkNameResponse carries the [out] parameters and return value of ApiGetResourceNetworkName.
type apiGetResourceNetworkNameResponse struct {
	LpszName   ndr.WSTR
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetResourceNetworkName calls ApiGetResourceNetworkName (opnum 112) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetResourceNetworkName(rpc ndr.Invoker, hResource mscmrp.HRES_RPC) (LpszName ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetResourceNetworkNameRequest{
		HResource: hResource,
	}
	var resp apiGetResourceNetworkNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetResourceNetworkName: %w", err)
		return
	}
	LpszName = resp.LpszName
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetResourceNetworkName failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
