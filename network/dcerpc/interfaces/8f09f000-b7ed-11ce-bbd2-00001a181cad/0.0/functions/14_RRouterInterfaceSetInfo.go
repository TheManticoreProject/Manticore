package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceSetInfoRequest carries the [in] parameters of RRouterInterfaceSetInfo.
type rRouterInterfaceSetInfoRequest struct {
	DwLevel     ndr.DWORD
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	HInterface  ndr.DWORD
}

func (*rRouterInterfaceSetInfoRequest) Opnum() uint16 { return dimsvc.OpnumRRouterInterfaceSetInfo }

// rRouterInterfaceSetInfoResponse carries the [out] parameters and return value of RRouterInterfaceSetInfo.
type rRouterInterfaceSetInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceSetInfo calls RRouterInterfaceSetInfo (opnum 14) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceSetInfo(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, hInterface ndr.DWORD) (err error) {
	req := &rRouterInterfaceSetInfoRequest{
		DwLevel:     dwLevel,
		PInfoStruct: pInfoStruct,
		HInterface:  hInterface,
	}
	var resp rRouterInterfaceSetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceSetInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceSetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
