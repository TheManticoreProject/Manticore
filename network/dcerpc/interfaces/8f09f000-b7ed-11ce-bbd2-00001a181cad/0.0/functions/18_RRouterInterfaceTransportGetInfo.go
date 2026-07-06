package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceTransportGetInfoRequest carries the [in] parameters of RRouterInterfaceTransportGetInfo.
type rRouterInterfaceTransportGetInfoRequest struct {
	HInterface    ndr.DWORD
	DwTransportId ndr.DWORD
	PInfoStruct   msrrasm.DIM_INTERFACE_CONTAINER
}

func (*rRouterInterfaceTransportGetInfoRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceTransportGetInfo
}

// rRouterInterfaceTransportGetInfoResponse carries the [out] parameters and return value of RRouterInterfaceTransportGetInfo.
type rRouterInterfaceTransportGetInfoResponse struct {
	PInfoStruct msrrasm.DIM_INTERFACE_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceTransportGetInfo calls RRouterInterfaceTransportGetInfo (opnum 18) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceTransportGetInfo(rpc ndr.Invoker, hInterface ndr.DWORD, dwTransportId ndr.DWORD, pInfoStruct msrrasm.DIM_INTERFACE_CONTAINER) (PInfoStruct msrrasm.DIM_INTERFACE_CONTAINER, err error) {
	req := &rRouterInterfaceTransportGetInfoRequest{
		HInterface:    hInterface,
		DwTransportId: dwTransportId,
		PInfoStruct:   pInfoStruct,
	}
	var resp rRouterInterfaceTransportGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceTransportGetInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceTransportGetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
