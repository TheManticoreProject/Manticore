package functions

import (
	"fmt"

	NetEventForwarder "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/22e5386d-8b12-4bf0-b0ec-6a1ea419e366/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslrec "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lrec"
)

// rpcNetEventOpenSessionRequest carries the [in] parameters of RpcNetEventOpenSession.
type rpcNetEventOpenSessionRequest struct {
	LoggerName ndr.WSTR
}

func (*rpcNetEventOpenSessionRequest) Opnum() uint16 {
	return NetEventForwarder.OpnumRpcNetEventOpenSession
}

// rpcNetEventOpenSessionResponse carries the [out] parameters and return value of RpcNetEventOpenSession.
type rpcNetEventOpenSessionResponse struct {
	SessionHandle mslrec.PSESSION_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcNetEventOpenSession calls RpcNetEventOpenSession (opnum 0) ([MS-LREC] 3.1.4.2.1),
// opening a context handle to the running event session named loggerName (the Name of a
// previously started MSFT_NetEventSession). On success the server begins accumulating
// matching events and returns the session handle.
func RpcNetEventOpenSession(rpc ndr.Invoker, loggerName ndr.WSTR) (SessionHandle mslrec.PSESSION_HANDLE, err error) {
	req := &rpcNetEventOpenSessionRequest{
		LoggerName: loggerName,
	}
	var resp rpcNetEventOpenSessionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcNetEventOpenSession: %w", err)
		return
	}
	SessionHandle = resp.SessionHandle
	if uint32(resp.Status) != NetEventForwarder.StatusSuccess {
		err = fmt.Errorf("RpcNetEventOpenSession failed: %s", NetEventForwarder.StatusString(uint32(resp.Status)))
	}
	return
}
