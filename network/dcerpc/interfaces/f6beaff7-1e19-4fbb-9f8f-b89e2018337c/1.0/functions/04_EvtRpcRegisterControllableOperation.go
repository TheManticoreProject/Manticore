package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcRegisterControllableOperationRequest carries the [in] parameters of EvtRpcRegisterControllableOperation.
type evtRpcRegisterControllableOperationRequest struct {
}

func (*evtRpcRegisterControllableOperationRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcRegisterControllableOperation
}

// evtRpcRegisterControllableOperationResponse carries the [out] parameters and return value of EvtRpcRegisterControllableOperation.
type evtRpcRegisterControllableOperationResponse struct {
	Handle mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcRegisterControllableOperation calls EvtRpcRegisterControllableOperation (opnum 4) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcRegisterControllableOperation(rpc ndr.Invoker) (Handle mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL, err error) {
	req := &evtRpcRegisterControllableOperationRequest{}
	var resp evtRpcRegisterControllableOperationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcRegisterControllableOperation: %w", err)
		return
	}
	Handle = resp.Handle
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcRegisterControllableOperation failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
