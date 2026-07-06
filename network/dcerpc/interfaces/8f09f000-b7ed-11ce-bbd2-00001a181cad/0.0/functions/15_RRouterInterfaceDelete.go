package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceDeleteRequest carries the [in] parameters of RRouterInterfaceDelete.
type rRouterInterfaceDeleteRequest struct {
	HInterface ndr.DWORD
}

func (*rRouterInterfaceDeleteRequest) Opnum() uint16 { return dimsvc.OpnumRRouterInterfaceDelete }

// rRouterInterfaceDeleteResponse carries the [out] parameters and return value of RRouterInterfaceDelete.
type rRouterInterfaceDeleteResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceDelete calls RRouterInterfaceDelete (opnum 15) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceDelete(rpc ndr.Invoker, hInterface ndr.DWORD) (err error) {
	req := &rRouterInterfaceDeleteRequest{
		HInterface: hInterface,
	}
	var resp rRouterInterfaceDeleteResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceDelete: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceDelete failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
