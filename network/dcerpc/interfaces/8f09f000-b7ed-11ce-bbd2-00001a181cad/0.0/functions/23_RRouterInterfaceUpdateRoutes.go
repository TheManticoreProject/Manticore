package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceUpdateRoutesRequest carries the [in] parameters of RRouterInterfaceUpdateRoutes.
type rRouterInterfaceUpdateRoutesRequest struct {
	HInterface        ndr.DWORD
	DwTransportId     ndr.DWORD
	HEvent            ndr.DWORD
	DwClientProcessId ndr.DWORD
}

func (*rRouterInterfaceUpdateRoutesRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceUpdateRoutes
}

// rRouterInterfaceUpdateRoutesResponse carries the [out] parameters and return value of RRouterInterfaceUpdateRoutes.
type rRouterInterfaceUpdateRoutesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceUpdateRoutes calls RRouterInterfaceUpdateRoutes (opnum 23) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceUpdateRoutes(rpc ndr.Invoker, hInterface ndr.DWORD, dwTransportId ndr.DWORD, hEvent ndr.DWORD, dwClientProcessId ndr.DWORD) (err error) {
	req := &rRouterInterfaceUpdateRoutesRequest{
		HInterface:        hInterface,
		DwTransportId:     dwTransportId,
		HEvent:            hEvent,
		DwClientProcessId: dwClientProcessId,
	}
	var resp rRouterInterfaceUpdateRoutesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceUpdateRoutes: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceUpdateRoutes failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
