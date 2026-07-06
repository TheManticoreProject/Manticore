package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceDisconnectRequest carries the [in] parameters of RRouterInterfaceDisconnect.
type rRouterInterfaceDisconnectRequest struct {
	HInterface ndr.DWORD
}

func (*rRouterInterfaceDisconnectRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceDisconnect
}

// rRouterInterfaceDisconnectResponse carries the [out] parameters and return value of RRouterInterfaceDisconnect.
type rRouterInterfaceDisconnectResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceDisconnect calls RRouterInterfaceDisconnect (opnum 22) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceDisconnect(rpc ndr.Invoker, hInterface ndr.DWORD) (err error) {
	req := &rRouterInterfaceDisconnectRequest{
		HInterface: hInterface,
	}
	var resp rRouterInterfaceDisconnectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceDisconnect: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceDisconnect failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
