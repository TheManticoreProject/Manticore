package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcQueryNextRequest carries the [in] parameters of EvtRpcQueryNext.
type evtRpcQueryNextRequest struct {
	LogQuery            mseven6.PCONTEXT_HANDLE_LOG_QUERY
	NumRequestedRecords ndr.DWORD
	TimeOutEnd          ndr.DWORD
	Flags               ndr.DWORD
}

func (*evtRpcQueryNextRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcQueryNext }

// evtRpcQueryNextResponse carries the [out] parameters and return value of EvtRpcQueryNext.
type evtRpcQueryNextResponse struct {
	NumActualRecords ndr.DWORD
	EventDataIndices []ndr.DWORD `ndr:"unique,size_is=NumActualRecords"`
	EventDataSizes   []ndr.DWORD `ndr:"unique,size_is=NumActualRecords"`
	ResultBufferSize ndr.DWORD
	ResultBuffer     []uint8   `ndr:"unique,size_is=ResultBufferSize"`
	Status           ndr.DWORD `ndr:"retval"`
}

// EvtRpcQueryNext calls EvtRpcQueryNext (opnum 11) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcQueryNext(rpc ndr.Invoker, logQuery mseven6.PCONTEXT_HANDLE_LOG_QUERY, numRequestedRecords ndr.DWORD, timeOutEnd ndr.DWORD, flags ndr.DWORD) (NumActualRecords ndr.DWORD, EventDataIndices []ndr.DWORD, EventDataSizes []ndr.DWORD, ResultBufferSize ndr.DWORD, ResultBuffer []uint8, err error) {
	req := &evtRpcQueryNextRequest{
		LogQuery:            logQuery,
		NumRequestedRecords: numRequestedRecords,
		TimeOutEnd:          timeOutEnd,
		Flags:               flags,
	}
	var resp evtRpcQueryNextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcQueryNext: %w", err)
		return
	}
	NumActualRecords = resp.NumActualRecords
	EventDataIndices = resp.EventDataIndices
	EventDataSizes = resp.EventDataSizes
	ResultBufferSize = resp.ResultBufferSize
	ResultBuffer = resp.ResultBuffer
	// A pull read reports end-of-stream and timeouts through the return code; both are
	// benign — the caller inspects NumActualRecords ([MS-EVEN6] 3.1.4.13).
	switch uint32(resp.Status) {
	case IEventService.StatusSuccess, IEventService.ErrorNoMoreItems, IEventService.ErrorTimeout:
	default:
		err = fmt.Errorf("EvtRpcQueryNext failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
