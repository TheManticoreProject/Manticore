package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_PurgeQueueRequest carries the [in] parameters of R_PurgeQueue.
type r_PurgeQueueRequest struct {
	PhContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
}

func (*r_PurgeQueueRequest) Opnum() uint16 { return RemoteRead.OpnumR_PurgeQueue }

// r_PurgeQueueResponse carries the [out] parameters and return value of R_PurgeQueue.
type r_PurgeQueueResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_PurgeQueue calls R_PurgeQueue (opnum 6) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_PurgeQueue(rpc ndr.Invoker, phContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE) (err error) {
	req := &r_PurgeQueueRequest{
		PhContext: phContext,
	}
	var resp r_PurgeQueueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_PurgeQueue: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_PurgeQueue failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
