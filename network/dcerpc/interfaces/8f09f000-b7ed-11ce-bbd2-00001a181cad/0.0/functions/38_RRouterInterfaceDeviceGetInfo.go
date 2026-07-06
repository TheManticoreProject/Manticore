package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceDeviceGetInfoRequest carries the [in] parameters of RRouterInterfaceDeviceGetInfo.
type rRouterInterfaceDeviceGetInfoRequest struct {
	DwLevel     ndr.DWORD
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	DwIndex     ndr.DWORD
	HInterface  ndr.DWORD
}

func (*rRouterInterfaceDeviceGetInfoRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceDeviceGetInfo
}

// rRouterInterfaceDeviceGetInfoResponse carries the [out] parameters and return value of RRouterInterfaceDeviceGetInfo.
type rRouterInterfaceDeviceGetInfoResponse struct {
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceDeviceGetInfo calls RRouterInterfaceDeviceGetInfo (opnum 38) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceDeviceGetInfo(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, dwIndex ndr.DWORD, hInterface ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, err error) {
	req := &rRouterInterfaceDeviceGetInfoRequest{
		DwLevel:     dwLevel,
		PInfoStruct: pInfoStruct,
		DwIndex:     dwIndex,
		HInterface:  hInterface,
	}
	var resp rRouterInterfaceDeviceGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceDeviceGetInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceDeviceGetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
