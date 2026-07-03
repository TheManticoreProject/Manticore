package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_MoveMessageRequest carries the [in] parameters of R_MoveMessage.
type r_MoveMessageRequest struct {
	PhContextFrom  msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
	UllContextTo   uint64
	LookupId       uint64
	PTransactionId msmqmq.XACTUOW
}

func (*r_MoveMessageRequest) Opnum() uint16 { return RemoteRead.OpnumR_MoveMessage }

// r_MoveMessageResponse carries the [out] parameters and return value of R_MoveMessage.
type r_MoveMessageResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_MoveMessage calls R_MoveMessage (opnum 10) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_MoveMessage(rpc ndr.Invoker, phContextFrom msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE, ullContextTo uint64, lookupId uint64, pTransactionId msmqmq.XACTUOW) (err error) {
	req := &r_MoveMessageRequest{
		PhContextFrom:  phContextFrom,
		UllContextTo:   ullContextTo,
		LookupId:       lookupId,
		PTransactionId: pTransactionId,
	}
	var resp r_MoveMessageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_MoveMessage: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_MoveMessage failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
