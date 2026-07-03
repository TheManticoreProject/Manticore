package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_CloseQueueRequest carries the [in] parameters of R_CloseQueue.
type r_CloseQueueRequest struct {
	PphContext msmqrr.QUEUE_CONTEXT_HANDLE_SERIALIZE
}

func (*r_CloseQueueRequest) Opnum() uint16 { return RemoteRead.OpnumR_CloseQueue }

// r_CloseQueueResponse carries the [out] parameters and return value of R_CloseQueue.
type r_CloseQueueResponse struct {
	PphContext msmqrr.QUEUE_CONTEXT_HANDLE_SERIALIZE
	Status     ndr.DWORD `ndr:"retval"`
}

// R_CloseQueue calls R_CloseQueue (opnum 3) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_CloseQueue(rpc ndr.Invoker, pphContext msmqrr.QUEUE_CONTEXT_HANDLE_SERIALIZE) (PphContext msmqrr.QUEUE_CONTEXT_HANDLE_SERIALIZE, err error) {
	req := &r_CloseQueueRequest{
		PphContext: pphContext,
	}
	var resp r_CloseQueueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_CloseQueue: %w", err)
		return
	}
	PphContext = resp.PphContext
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_CloseQueue failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
