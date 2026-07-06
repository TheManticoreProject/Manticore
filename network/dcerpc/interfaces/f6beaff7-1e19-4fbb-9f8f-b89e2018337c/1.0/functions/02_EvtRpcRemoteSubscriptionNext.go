package functions

// IDL source: [MS-EVEN6] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even6/2d808edd-719a-4c69-b34a-df766adb5f0c
// A fetched copy is kept at ms-even6.idl in the interface directory.

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcRemoteSubscriptionNextRequest carries the [in] parameters of EvtRpcRemoteSubscriptionNext.
type evtRpcRemoteSubscriptionNextRequest struct {
	Handle              mseven6.PCONTEXT_HANDLE_REMOTE_SUBSCRIPTION
	NumRequestedRecords ndr.DWORD
	TimeOut             ndr.DWORD
	Flags               ndr.DWORD
}

func (*evtRpcRemoteSubscriptionNextRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcRemoteSubscriptionNext
}

// evtRpcRemoteSubscriptionNextResponse carries the [out] parameters and return value of EvtRpcRemoteSubscriptionNext.
type evtRpcRemoteSubscriptionNextResponse struct {
	NumActualRecords ndr.DWORD
	EventDataIndices []ndr.DWORD `ndr:"unique,size_is=NumActualRecords"`
	EventDataSizes   []ndr.DWORD `ndr:"unique,size_is=NumActualRecords"`
	ResultBufferSize ndr.DWORD
	ResultBuffer     []uint8   `ndr:"unique,size_is=ResultBufferSize"`
	Status           ndr.DWORD `ndr:"retval"`
}

// EvtRpcRemoteSubscriptionNext calls EvtRpcRemoteSubscriptionNext (opnum 2) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcRemoteSubscriptionNext(rpc ndr.Invoker, handle mseven6.PCONTEXT_HANDLE_REMOTE_SUBSCRIPTION, numRequestedRecords ndr.DWORD, timeOut ndr.DWORD, flags ndr.DWORD) (NumActualRecords ndr.DWORD, EventDataIndices []ndr.DWORD, EventDataSizes []ndr.DWORD, ResultBufferSize ndr.DWORD, ResultBuffer []uint8, err error) {
	req := &evtRpcRemoteSubscriptionNextRequest{
		Handle:              handle,
		NumRequestedRecords: numRequestedRecords,
		TimeOut:             timeOut,
		Flags:               flags,
	}
	var resp evtRpcRemoteSubscriptionNextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcRemoteSubscriptionNext: %w", err)
		return
	}
	NumActualRecords = resp.NumActualRecords
	EventDataIndices = resp.EventDataIndices
	EventDataSizes = resp.EventDataSizes
	ResultBufferSize = resp.ResultBufferSize
	ResultBuffer = resp.ResultBuffer
	// A pull read reports end-of-stream and timeouts through the return code; both are
	// benign — the caller inspects NumActualRecords ([MS-EVEN6] 3.1.4.11).
	switch uint32(resp.Status) {
	case IEventService.StatusSuccess, IEventService.ErrorNoMoreItems, IEventService.ErrorTimeout:
	default:
		err = fmt.Errorf("EvtRpcRemoteSubscriptionNext failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
