package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceTransportGetGlobalInfoRequest carries the [in] parameters of RRouterInterfaceTransportGetGlobalInfo.
type rRouterInterfaceTransportGetGlobalInfoRequest struct {
	DwTransportId ndr.DWORD
	PInfoStruct   msrrasm.DIM_INTERFACE_CONTAINER
}

func (*rRouterInterfaceTransportGetGlobalInfoRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceTransportGetGlobalInfo
}

// rRouterInterfaceTransportGetGlobalInfoResponse carries the [out] parameters and return value of RRouterInterfaceTransportGetGlobalInfo.
type rRouterInterfaceTransportGetGlobalInfoResponse struct {
	PInfoStruct msrrasm.DIM_INTERFACE_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceTransportGetGlobalInfo calls RRouterInterfaceTransportGetGlobalInfo (opnum 10) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceTransportGetGlobalInfo(rpc ndr.Invoker, dwTransportId ndr.DWORD, pInfoStruct msrrasm.DIM_INTERFACE_CONTAINER) (PInfoStruct msrrasm.DIM_INTERFACE_CONTAINER, err error) {
	req := &rRouterInterfaceTransportGetGlobalInfoRequest{
		DwTransportId: dwTransportId,
		PInfoStruct:   pInfoStruct,
	}
	var resp rRouterInterfaceTransportGetGlobalInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceTransportGetGlobalInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceTransportGetGlobalInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
