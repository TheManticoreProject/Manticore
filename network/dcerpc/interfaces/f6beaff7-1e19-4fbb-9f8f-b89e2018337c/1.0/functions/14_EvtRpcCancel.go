package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcCancelRequest carries the [in] parameters of EvtRpcCancel.
type evtRpcCancelRequest struct {
	Handle mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL
}

func (*evtRpcCancelRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcCancel }

// evtRpcCancelResponse carries the [out] parameters and return value of EvtRpcCancel.
type evtRpcCancelResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcCancel calls EvtRpcCancel (opnum 14) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcCancel(rpc ndr.Invoker, handle mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL) (err error) {
	req := &evtRpcCancelRequest{
		Handle: handle,
	}
	var resp evtRpcCancelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcCancel: %w", err)
		return
	}
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcCancel failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
