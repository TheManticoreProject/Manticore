package functions

import (
	"fmt"

	NetEventForwarder "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/22e5386d-8b12-4bf0-b0ec-6a1ea419e366/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslrec "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lrec"
)

// rpcNetEventCloseSessionRequest carries the [in,out] parameter of RpcNetEventCloseSession:
// the session context handle to close.
type rpcNetEventCloseSessionRequest struct {
	SessionHandle mslrec.PSESSION_HANDLE
}

func (*rpcNetEventCloseSessionRequest) Opnum() uint16 {
	return NetEventForwarder.OpnumRpcNetEventCloseSession
}

// rpcNetEventCloseSessionResponse carries the [in,out] parameter of RpcNetEventCloseSession.
// The method is declared `void` in the IDL, so there is no return value on the wire —
// the response contains only the (server-zeroed) context handle.
type rpcNetEventCloseSessionResponse struct {
	SessionHandle mslrec.PSESSION_HANDLE
}

// RpcNetEventCloseSession calls RpcNetEventCloseSession (opnum 2) ([MS-LREC] 3.1.4.2.3),
// closing the event session referenced by sessionHandle. The method returns void; on
// success the server returns a zeroed context handle, which is returned to the caller.
func RpcNetEventCloseSession(rpc ndr.Invoker, sessionHandle mslrec.PSESSION_HANDLE) (mslrec.PSESSION_HANDLE, error) {
	req := &rpcNetEventCloseSessionRequest{
		SessionHandle: sessionHandle,
	}
	var resp rpcNetEventCloseSessionResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mslrec.PSESSION_HANDLE{}, fmt.Errorf("RpcNetEventCloseSession: %w", err)
	}
	return resp.SessionHandle, nil
}
