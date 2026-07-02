package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcOpenLogHandleRequest carries the [in] parameters of EvtRpcOpenLogHandle.
type evtRpcOpenLogHandleRequest struct {
	Channel ndr.WSTR
	Flags   ndr.DWORD
}

func (*evtRpcOpenLogHandleRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcOpenLogHandle }

// evtRpcOpenLogHandleResponse carries the [out] parameters and return value of EvtRpcOpenLogHandle.
type evtRpcOpenLogHandleResponse struct {
	Handle mseven6.PCONTEXT_HANDLE_LOG_HANDLE
	Error  mseven6.RpcInfo
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcOpenLogHandle calls EvtRpcOpenLogHandle (opnum 17) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcOpenLogHandle(rpc ndr.Invoker, channel ndr.WSTR, flags ndr.DWORD) (Handle mseven6.PCONTEXT_HANDLE_LOG_HANDLE, Error mseven6.RpcInfo, err error) {
	req := &evtRpcOpenLogHandleRequest{
		Channel: channel,
		Flags:   flags,
	}
	var resp evtRpcOpenLogHandleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcOpenLogHandle: %w", err)
		return
	}
	Handle = resp.Handle
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcOpenLogHandle failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
