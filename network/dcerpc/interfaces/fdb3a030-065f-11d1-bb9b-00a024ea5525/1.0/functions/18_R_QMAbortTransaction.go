package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// r_QMAbortTransactionRequest carries the [in] parameters of R_QMAbortTransaction.
type r_QMAbortTransactionRequest struct {
	PhIntXact msmqmp.RPC_INT_XACT_HANDLE
}

func (*r_QMAbortTransactionRequest) Opnum() uint16 { return qmcomm.OpnumR_QMAbortTransaction }

// r_QMAbortTransactionResponse carries the [out] parameters and return value of R_QMAbortTransaction.
type r_QMAbortTransactionResponse struct {
	PhIntXact msmqmp.RPC_INT_XACT_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// R_QMAbortTransaction calls R_QMAbortTransaction (opnum 18) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMAbortTransaction(rpc ndr.Invoker, phIntXact msmqmp.RPC_INT_XACT_HANDLE) (PhIntXact msmqmp.RPC_INT_XACT_HANDLE, err error) {
	req := &r_QMAbortTransactionRequest{
		PhIntXact: phIntXact,
	}
	var resp r_QMAbortTransactionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMAbortTransaction: %w", err)
		return
	}
	PhIntXact = resp.PhIntXact
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMAbortTransaction failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
