package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceTransportSetInfoRequest carries the [in] parameters of RRouterInterfaceTransportSetInfo.
type rRouterInterfaceTransportSetInfoRequest struct {
	HInterface    ndr.DWORD
	DwTransportId ndr.DWORD
	PInfoStruct   msrrasm.DIM_INTERFACE_CONTAINER
}

func (*rRouterInterfaceTransportSetInfoRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceTransportSetInfo
}

// rRouterInterfaceTransportSetInfoResponse carries the [out] parameters and return value of RRouterInterfaceTransportSetInfo.
type rRouterInterfaceTransportSetInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceTransportSetInfo calls RRouterInterfaceTransportSetInfo (opnum 19) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceTransportSetInfo(rpc ndr.Invoker, hInterface ndr.DWORD, dwTransportId ndr.DWORD, pInfoStruct msrrasm.DIM_INTERFACE_CONTAINER) (err error) {
	req := &rRouterInterfaceTransportSetInfoRequest{
		HInterface:    hInterface,
		DwTransportId: dwTransportId,
		PInfoStruct:   pInfoStruct,
	}
	var resp rRouterInterfaceTransportSetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceTransportSetInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceTransportSetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
