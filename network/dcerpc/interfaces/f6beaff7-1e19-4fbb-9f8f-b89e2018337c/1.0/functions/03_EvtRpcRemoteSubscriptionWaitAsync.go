package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcRemoteSubscriptionWaitAsyncRequest carries the [in] parameters of EvtRpcRemoteSubscriptionWaitAsync.
type evtRpcRemoteSubscriptionWaitAsyncRequest struct {
	Handle mseven6.PCONTEXT_HANDLE_REMOTE_SUBSCRIPTION
}

func (*evtRpcRemoteSubscriptionWaitAsyncRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcRemoteSubscriptionWaitAsync
}

// evtRpcRemoteSubscriptionWaitAsyncResponse carries the [out] parameters and return value of EvtRpcRemoteSubscriptionWaitAsync.
type evtRpcRemoteSubscriptionWaitAsyncResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcRemoteSubscriptionWaitAsync calls EvtRpcRemoteSubscriptionWaitAsync (opnum 3) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcRemoteSubscriptionWaitAsync(rpc ndr.Invoker, handle mseven6.PCONTEXT_HANDLE_REMOTE_SUBSCRIPTION) (err error) {
	req := &evtRpcRemoteSubscriptionWaitAsyncRequest{
		Handle: handle,
	}
	var resp evtRpcRemoteSubscriptionWaitAsyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcRemoteSubscriptionWaitAsync: %w", err)
		return
	}
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcRemoteSubscriptionWaitAsync failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
