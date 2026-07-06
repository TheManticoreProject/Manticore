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

// evtRpcCloseRequest carries the [in] parameter of EvtRpcClose.
type evtRpcCloseRequest struct {
	Handle mseven6.CONTEXT_HANDLE
}

func (*evtRpcCloseRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcClose }

// evtRpcCloseResponse carries the [out] parameter and return value of EvtRpcClose. On
// success the server returns the handle nulled out.
type evtRpcCloseResponse struct {
	Handle mseven6.CONTEXT_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcClose calls EvtRpcClose (opnum 13) ([MS-EVEN6] 3.1.4.35). It closes the context
// handle and returns the server's (nulled) handle value.
func EvtRpcClose(rpc ndr.Invoker, handle mseven6.CONTEXT_HANDLE) (mseven6.CONTEXT_HANDLE, error) {
	req := &evtRpcCloseRequest{
		Handle: handle,
	}
	var resp evtRpcCloseResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mseven6.CONTEXT_HANDLE{}, fmt.Errorf("EvtRpcClose: %w", err)
	}
	if uint32(resp.Status) != IEventService.StatusSuccess {
		return resp.Handle, fmt.Errorf("EvtRpcClose failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return resp.Handle, nil
}
