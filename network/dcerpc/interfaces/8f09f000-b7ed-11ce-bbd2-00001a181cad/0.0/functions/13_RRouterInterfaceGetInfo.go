package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceGetInfoRequest carries the [in] parameters of RRouterInterfaceGetInfo.
type rRouterInterfaceGetInfoRequest struct {
	DwLevel     ndr.DWORD
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	HInterface  ndr.DWORD
}

func (*rRouterInterfaceGetInfoRequest) Opnum() uint16 { return dimsvc.OpnumRRouterInterfaceGetInfo }

// rRouterInterfaceGetInfoResponse carries the [out] parameters and return value of RRouterInterfaceGetInfo.
type rRouterInterfaceGetInfoResponse struct {
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceGetInfo calls RRouterInterfaceGetInfo (opnum 13) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceGetInfo(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, hInterface ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, err error) {
	req := &rRouterInterfaceGetInfoRequest{
		DwLevel:     dwLevel,
		PInfoStruct: pInfoStruct,
		HInterface:  hInterface,
	}
	var resp rRouterInterfaceGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceGetInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceGetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
