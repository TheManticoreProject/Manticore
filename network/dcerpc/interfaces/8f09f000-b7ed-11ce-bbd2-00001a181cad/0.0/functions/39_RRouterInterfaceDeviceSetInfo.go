package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceDeviceSetInfoRequest carries the [in] parameters of RRouterInterfaceDeviceSetInfo.
type rRouterInterfaceDeviceSetInfoRequest struct {
	DwLevel     ndr.DWORD
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	DwIndex     ndr.DWORD
	HInterface  ndr.DWORD
}

func (*rRouterInterfaceDeviceSetInfoRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceDeviceSetInfo
}

// rRouterInterfaceDeviceSetInfoResponse carries the [out] parameters and return value of RRouterInterfaceDeviceSetInfo.
type rRouterInterfaceDeviceSetInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceDeviceSetInfo calls RRouterInterfaceDeviceSetInfo (opnum 39) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceDeviceSetInfo(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, dwIndex ndr.DWORD, hInterface ndr.DWORD) (err error) {
	req := &rRouterInterfaceDeviceSetInfoRequest{
		DwLevel:     dwLevel,
		PInfoStruct: pInfoStruct,
		DwIndex:     dwIndex,
		HInterface:  hInterface,
	}
	var resp rRouterInterfaceDeviceSetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceDeviceSetInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceDeviceSetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
