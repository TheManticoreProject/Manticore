package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceTransportRemoveRequest carries the [in] parameters of RRouterInterfaceTransportRemove.
type rRouterInterfaceTransportRemoveRequest struct {
	HInterface    ndr.DWORD
	DwTransportId ndr.DWORD
}

func (*rRouterInterfaceTransportRemoveRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceTransportRemove
}

// rRouterInterfaceTransportRemoveResponse carries the [out] parameters and return value of RRouterInterfaceTransportRemove.
type rRouterInterfaceTransportRemoveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceTransportRemove calls RRouterInterfaceTransportRemove (opnum 16) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceTransportRemove(rpc ndr.Invoker, hInterface ndr.DWORD, dwTransportId ndr.DWORD) (err error) {
	req := &rRouterInterfaceTransportRemoveRequest{
		HInterface:    hInterface,
		DwTransportId: dwTransportId,
	}
	var resp rRouterInterfaceTransportRemoveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceTransportRemove: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceTransportRemove failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
