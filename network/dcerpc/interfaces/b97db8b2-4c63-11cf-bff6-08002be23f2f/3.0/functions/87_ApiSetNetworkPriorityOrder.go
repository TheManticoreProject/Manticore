package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// apiSetNetworkPriorityOrderRequest carries the [in] parameters of ApiSetNetworkPriorityOrder.
type apiSetNetworkPriorityOrderRequest struct {
	NetworkCount  ndr.DWORD
	NetworkIdList []*ndr.WSTR `ndr:"elem=unique,ref,size_is=NetworkCount"`
}

func (*apiSetNetworkPriorityOrderRequest) Opnum() uint16 {
	return clusapi.OpnumApiSetNetworkPriorityOrder
}

// apiSetNetworkPriorityOrderResponse carries the [out] parameters and return value of ApiSetNetworkPriorityOrder.
type apiSetNetworkPriorityOrderResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetNetworkPriorityOrder calls ApiSetNetworkPriorityOrder (opnum 87) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetNetworkPriorityOrder(rpc ndr.Invoker, networkCount ndr.DWORD, networkIdList []*ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetNetworkPriorityOrderRequest{
		NetworkCount:  networkCount,
		NetworkIdList: networkIdList,
	}
	var resp apiSetNetworkPriorityOrderResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetNetworkPriorityOrder: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetNetworkPriorityOrder failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
