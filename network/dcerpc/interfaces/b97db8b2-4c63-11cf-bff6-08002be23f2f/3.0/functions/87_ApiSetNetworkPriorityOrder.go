package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

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
