package functions

import (
	"fmt"

	NetEventForwarder "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/22e5386d-8b12-4bf0-b0ec-6a1ea419e366/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslrec "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lrec"
)

// rpcNetEventReceiveDataRequest carries the [in] parameters of RpcNetEventReceiveData.
type rpcNetEventReceiveDataRequest struct {
	SessionHandle mslrec.PSESSION_HANDLE
}

func (*rpcNetEventReceiveDataRequest) Opnum() uint16 {
	return NetEventForwarder.OpnumRpcNetEventReceiveData
}

// rpcNetEventReceiveDataResponse carries the [out] parameters and return value of RpcNetEventReceiveData.
type rpcNetEventReceiveDataResponse struct {
	EventBuffer mslrec.EVENT_BUFFER
	Status      ndr.DWORD `ndr:"retval"`
}

// RpcNetEventReceiveData calls RpcNetEventReceiveData (opnum 1) ([MS-LREC] 3.1.4.2.2),
// draining buffered events for the session referenced by sessionHandle. The returned
// EVENT_BUFFER holds one or more NET_EVENT_DATA_HEADER structures with their payloads;
// the server determines the buffer size and may block until enough events accumulate.
func RpcNetEventReceiveData(rpc ndr.Invoker, sessionHandle mslrec.PSESSION_HANDLE) (EventBuffer mslrec.EVENT_BUFFER, err error) {
	req := &rpcNetEventReceiveDataRequest{
		SessionHandle: sessionHandle,
	}
	var resp rpcNetEventReceiveDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcNetEventReceiveData: %w", err)
		return
	}
	EventBuffer = resp.EventBuffer
	if uint32(resp.Status) != NetEventForwarder.StatusSuccess {
		err = fmt.Errorf("RpcNetEventReceiveData failed: %s", NetEventForwarder.StatusString(uint32(resp.Status)))
	}
	return
}
