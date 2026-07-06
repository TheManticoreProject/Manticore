package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceConnectRequest carries the [in] parameters of RRouterInterfaceConnect.
type rRouterInterfaceConnectRequest struct {
	HInterface         ndr.DWORD
	HEvent             ndr.DWORD
	FBlocking          ndr.DWORD
	DwCallersProcessId ndr.DWORD
}

func (*rRouterInterfaceConnectRequest) Opnum() uint16 { return dimsvc.OpnumRRouterInterfaceConnect }

// rRouterInterfaceConnectResponse carries the [out] parameters and return value of RRouterInterfaceConnect.
type rRouterInterfaceConnectResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceConnect calls RRouterInterfaceConnect (opnum 21) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceConnect(rpc ndr.Invoker, hInterface ndr.DWORD, hEvent ndr.DWORD, fBlocking ndr.DWORD, dwCallersProcessId ndr.DWORD) (err error) {
	req := &rRouterInterfaceConnectRequest{
		HInterface:         hInterface,
		HEvent:             hEvent,
		FBlocking:          fBlocking,
		DwCallersProcessId: dwCallersProcessId,
	}
	var resp rRouterInterfaceConnectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceConnect: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceConnect failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
