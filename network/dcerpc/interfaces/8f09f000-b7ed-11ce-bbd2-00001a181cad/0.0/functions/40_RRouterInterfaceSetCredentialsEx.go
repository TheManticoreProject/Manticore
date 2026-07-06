package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceSetCredentialsExRequest carries the [in] parameters of RRouterInterfaceSetCredentialsEx.
type rRouterInterfaceSetCredentialsExRequest struct {
	DwLevel     ndr.DWORD
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	HInterface  ndr.DWORD
}

func (*rRouterInterfaceSetCredentialsExRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceSetCredentialsEx
}

// rRouterInterfaceSetCredentialsExResponse carries the [out] parameters and return value of RRouterInterfaceSetCredentialsEx.
type rRouterInterfaceSetCredentialsExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceSetCredentialsEx calls RRouterInterfaceSetCredentialsEx (opnum 40) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceSetCredentialsEx(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, hInterface ndr.DWORD) (err error) {
	req := &rRouterInterfaceSetCredentialsExRequest{
		DwLevel:     dwLevel,
		PInfoStruct: pInfoStruct,
		HInterface:  hInterface,
	}
	var resp rRouterInterfaceSetCredentialsExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceSetCredentialsEx: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceSetCredentialsEx failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
