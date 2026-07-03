package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// rpc_ACCloseHandleRequest carries the [in] parameters of rpc_ACCloseHandle.
type rpc_ACCloseHandleRequest struct {
	PhQueue msmqmp.RPC_QUEUE_HANDLE
}

func (*rpc_ACCloseHandleRequest) Opnum() uint16 { return qmcomm.Opnumrpc_ACCloseHandle }

// rpc_ACCloseHandleResponse carries the [out] parameters and return value of rpc_ACCloseHandle.
type rpc_ACCloseHandleResponse struct {
	PhQueue msmqmp.RPC_QUEUE_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// Rpc_ACCloseHandle calls rpc_ACCloseHandle (opnum 20) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func Rpc_ACCloseHandle(rpc ndr.Invoker, phQueue msmqmp.RPC_QUEUE_HANDLE) (PhQueue msmqmp.RPC_QUEUE_HANDLE, err error) {
	req := &rpc_ACCloseHandleRequest{
		PhQueue: phQueue,
	}
	var resp rpc_ACCloseHandleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("rpc_ACCloseHandle: %w", err)
		return
	}
	PhQueue = resp.PhQueue
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("rpc_ACCloseHandle failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
