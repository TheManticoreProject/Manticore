package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetNetworkIdRequest carries the [in] parameters of ApiGetNetworkId.
type apiGetNetworkIdRequest struct {
	HNetwork mscmrp.HNETWORK_RPC
}

func (*apiGetNetworkIdRequest) Opnum() uint16 { return clusapi.OpnumApiGetNetworkId }

// apiGetNetworkIdResponse carries the [out] parameters and return value of ApiGetNetworkId.
type apiGetNetworkIdResponse struct {
	PGuid      ndr.WSTR
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetNetworkId calls ApiGetNetworkId (opnum 86) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNetworkId(rpc ndr.Invoker, hNetwork mscmrp.HNETWORK_RPC) (PGuid ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetNetworkIdRequest{
		HNetwork: hNetwork,
	}
	var resp apiGetNetworkIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNetworkId: %w", err)
		return
	}
	PGuid = resp.PGuid
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNetworkId failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
