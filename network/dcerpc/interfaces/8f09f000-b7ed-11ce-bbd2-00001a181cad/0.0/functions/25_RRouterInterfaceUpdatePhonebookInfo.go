package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceUpdatePhonebookInfoRequest carries the [in] parameters of RRouterInterfaceUpdatePhonebookInfo.
type rRouterInterfaceUpdatePhonebookInfoRequest struct {
	HInterface ndr.DWORD
}

func (*rRouterInterfaceUpdatePhonebookInfoRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceUpdatePhonebookInfo
}

// rRouterInterfaceUpdatePhonebookInfoResponse carries the [out] parameters and return value of RRouterInterfaceUpdatePhonebookInfo.
type rRouterInterfaceUpdatePhonebookInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceUpdatePhonebookInfo calls RRouterInterfaceUpdatePhonebookInfo (opnum 25) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceUpdatePhonebookInfo(rpc ndr.Invoker, hInterface ndr.DWORD) (err error) {
	req := &rRouterInterfaceUpdatePhonebookInfoRequest{
		HInterface: hInterface,
	}
	var resp rRouterInterfaceUpdatePhonebookInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceUpdatePhonebookInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceUpdatePhonebookInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
