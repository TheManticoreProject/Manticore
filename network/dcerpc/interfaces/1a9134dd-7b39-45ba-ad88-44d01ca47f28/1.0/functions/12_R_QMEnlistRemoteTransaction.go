package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// r_QMEnlistRemoteTransactionRequest carries the [in] parameters of R_QMEnlistRemoteTransaction.
type r_QMEnlistRemoteTransactionRequest struct {
	PTransactionId     msmqmq.XACTUOW
	CbPropagationToken ndr.DWORD
	PbPropagationToken []uint8 `ndr:"ref,size_is=CbPropagationToken"`
	PQueueFormat       msmqmq.QUEUE_FORMAT
}

func (*r_QMEnlistRemoteTransactionRequest) Opnum() uint16 {
	return RemoteRead.OpnumR_QMEnlistRemoteTransaction
}

// r_QMEnlistRemoteTransactionResponse carries the [out] parameters and return value of R_QMEnlistRemoteTransaction.
type r_QMEnlistRemoteTransactionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_QMEnlistRemoteTransaction calls R_QMEnlistRemoteTransaction (opnum 12) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_QMEnlistRemoteTransaction(rpc ndr.Invoker, pTransactionId msmqmq.XACTUOW, cbPropagationToken ndr.DWORD, pbPropagationToken []uint8, pQueueFormat msmqmq.QUEUE_FORMAT) (err error) {
	req := &r_QMEnlistRemoteTransactionRequest{
		PTransactionId:     pTransactionId,
		CbPropagationToken: cbPropagationToken,
		PbPropagationToken: pbPropagationToken,
		PQueueFormat:       pQueueFormat,
	}
	var resp r_QMEnlistRemoteTransactionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMEnlistRemoteTransaction: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_QMEnlistRemoteTransaction failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
