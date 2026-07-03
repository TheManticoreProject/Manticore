package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// rpc_ACPurgeQueueRequest carries the [in] parameters of rpc_ACPurgeQueue.
type rpc_ACPurgeQueueRequest struct {
	HQueue msmqmp.RPC_QUEUE_HANDLE
}

func (*rpc_ACPurgeQueueRequest) Opnum() uint16 { return qmcomm.Opnumrpc_ACPurgeQueue }

// rpc_ACPurgeQueueResponse carries the [out] parameters and return value of rpc_ACPurgeQueue.
type rpc_ACPurgeQueueResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Rpc_ACPurgeQueue calls rpc_ACPurgeQueue (opnum 27) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func Rpc_ACPurgeQueue(rpc ndr.Invoker, hQueue msmqmp.RPC_QUEUE_HANDLE) (err error) {
	req := &rpc_ACPurgeQueueRequest{
		HQueue: hQueue,
	}
	var resp rpc_ACPurgeQueueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("rpc_ACPurgeQueue: %w", err)
		return
	}
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("rpc_ACPurgeQueue failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
