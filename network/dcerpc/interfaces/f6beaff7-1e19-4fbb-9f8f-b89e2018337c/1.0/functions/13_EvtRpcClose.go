package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ContextHandle is the 20-octet RPC context handle ([MS-RPCE] 2.3.2.2) that EvtRpcClose
// closes. The IDL types it as a generic [in, out] context_handle void**, so it accepts any
// of the interface's handles (log query, log handle, operation control, subscription, …);
// callers convert their typed [20]byte handle to this type. It is transmitted inline (no
// referent id), matching every other context handle in this interface.
type ContextHandle [20]byte

// evtRpcCloseRequest carries the [in] parameter of EvtRpcClose.
type evtRpcCloseRequest struct {
	Handle ContextHandle
}

func (*evtRpcCloseRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcClose }

// evtRpcCloseResponse carries the [out] parameter and return value of EvtRpcClose. On
// success the server returns the handle nulled out.
type evtRpcCloseResponse struct {
	Handle ContextHandle
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcClose calls EvtRpcClose (opnum 13) ([MS-EVEN6] 3.1.4.35). It closes the context
// handle and returns the server's (nulled) handle value.
func EvtRpcClose(rpc ndr.Invoker, handle ContextHandle) (ContextHandle, error) {
	req := &evtRpcCloseRequest{
		Handle: handle,
	}
	var resp evtRpcCloseResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return ContextHandle{}, fmt.Errorf("EvtRpcClose: %w", err)
	}
	if uint32(resp.Status) != IEventService.StatusSuccess {
		return resp.Handle, fmt.Errorf("EvtRpcClose failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
