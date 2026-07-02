package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcRemoteSubscriptionNextAsyncRequest carries the [in] parameters of EvtRpcRemoteSubscriptionNextAsync.
type evtRpcRemoteSubscriptionNextAsyncRequest struct {
	Handle              mseven6.PCONTEXT_HANDLE_REMOTE_SUBSCRIPTION
	NumRequestedRecords ndr.DWORD
	Flags               ndr.DWORD
}

func (*evtRpcRemoteSubscriptionNextAsyncRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcRemoteSubscriptionNextAsync
}

// evtRpcRemoteSubscriptionNextAsyncResponse carries the [out] parameters and return value of EvtRpcRemoteSubscriptionNextAsync.
type evtRpcRemoteSubscriptionNextAsyncResponse struct {
	NumActualRecords ndr.DWORD
	EventDataIndices []ndr.DWORD `ndr:"unique,size_is=NumActualRecords"`
	EventDataSizes   []ndr.DWORD `ndr:"unique,size_is=NumActualRecords"`
	ResultBufferSize ndr.DWORD
	ResultBuffer     []uint8   `ndr:"unique,size_is=ResultBufferSize"`
	Status           ndr.DWORD `ndr:"retval"`
}

// EvtRpcRemoteSubscriptionNextAsync calls EvtRpcRemoteSubscriptionNextAsync (opnum 1) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcRemoteSubscriptionNextAsync(rpc ndr.Invoker, handle mseven6.PCONTEXT_HANDLE_REMOTE_SUBSCRIPTION, numRequestedRecords ndr.DWORD, flags ndr.DWORD) (NumActualRecords ndr.DWORD, EventDataIndices []ndr.DWORD, EventDataSizes []ndr.DWORD, ResultBufferSize ndr.DWORD, ResultBuffer []uint8, err error) {
	req := &evtRpcRemoteSubscriptionNextAsyncRequest{
		Handle:              handle,
		NumRequestedRecords: numRequestedRecords,
		Flags:               flags,
	}
	var resp evtRpcRemoteSubscriptionNextAsyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcRemoteSubscriptionNextAsync: %w", err)
		return
	}
	NumActualRecords = resp.NumActualRecords
	EventDataIndices = resp.EventDataIndices
	EventDataSizes = resp.EventDataSizes
	ResultBufferSize = resp.ResultBufferSize
	ResultBuffer = resp.ResultBuffer
	// An async pull read reports an exhausted stream through the return code; it is benign
	// — the caller inspects NumActualRecords ([MS-EVEN6] 3.1.4.10).
	switch uint32(resp.Status) {
	case IEventService.StatusSuccess, IEventService.ErrorNoMoreItems:
	default:
		err = fmt.Errorf("EvtRpcRemoteSubscriptionNextAsync failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
