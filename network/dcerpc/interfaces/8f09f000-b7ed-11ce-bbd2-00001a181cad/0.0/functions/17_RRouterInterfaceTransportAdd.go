package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceTransportAddRequest carries the [in] parameters of RRouterInterfaceTransportAdd.
type rRouterInterfaceTransportAddRequest struct {
	HInterface    ndr.DWORD
	DwTransportId ndr.DWORD
	PInfoStruct   msrrasm.DIM_INTERFACE_CONTAINER
}

func (*rRouterInterfaceTransportAddRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceTransportAdd
}

// rRouterInterfaceTransportAddResponse carries the [out] parameters and return value of RRouterInterfaceTransportAdd.
type rRouterInterfaceTransportAddResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceTransportAdd calls RRouterInterfaceTransportAdd (opnum 17) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceTransportAdd(rpc ndr.Invoker, hInterface ndr.DWORD, dwTransportId ndr.DWORD, pInfoStruct msrrasm.DIM_INTERFACE_CONTAINER) (err error) {
	req := &rRouterInterfaceTransportAddRequest{
		HInterface:    hInterface,
		DwTransportId: dwTransportId,
		PInfoStruct:   pInfoStruct,
	}
	var resp rRouterInterfaceTransportAddResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceTransportAdd: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceTransportAdd failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
