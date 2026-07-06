package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceTransportSetGlobalInfoRequest carries the [in] parameters of RRouterInterfaceTransportSetGlobalInfo.
type rRouterInterfaceTransportSetGlobalInfoRequest struct {
	DwTransportId ndr.DWORD
	PInfoStruct   msrrasm.DIM_INTERFACE_CONTAINER
}

func (*rRouterInterfaceTransportSetGlobalInfoRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceTransportSetGlobalInfo
}

// rRouterInterfaceTransportSetGlobalInfoResponse carries the [out] parameters and return value of RRouterInterfaceTransportSetGlobalInfo.
type rRouterInterfaceTransportSetGlobalInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceTransportSetGlobalInfo calls RRouterInterfaceTransportSetGlobalInfo (opnum 9) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceTransportSetGlobalInfo(rpc ndr.Invoker, dwTransportId ndr.DWORD, pInfoStruct msrrasm.DIM_INTERFACE_CONTAINER) (err error) {
	req := &rRouterInterfaceTransportSetGlobalInfoRequest{
		DwTransportId: dwTransportId,
		PInfoStruct:   pInfoStruct,
	}
	var resp rRouterInterfaceTransportSetGlobalInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceTransportSetGlobalInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceTransportSetGlobalInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
