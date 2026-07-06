package functions

// IDL source: [MS-MQMP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmp/a54c09de-1d72-47f0-9184-d7e5046b2ba1
// A fetched copy is kept at ms-mqmp.idl in the interface directory.

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// r_QMEnlistInternalTransactionRequest carries the [in] parameters of R_QMEnlistInternalTransaction.
type r_QMEnlistInternalTransactionRequest struct {
	PUow msmqmq.XACTUOW
}

func (*r_QMEnlistInternalTransactionRequest) Opnum() uint16 {
	return qmcomm.OpnumR_QMEnlistInternalTransaction
}

// r_QMEnlistInternalTransactionResponse carries the [out] parameters and return value of R_QMEnlistInternalTransaction.
type r_QMEnlistInternalTransactionResponse struct {
	PhIntXact msmqmp.RPC_INT_XACT_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// R_QMEnlistInternalTransaction calls R_QMEnlistInternalTransaction (opnum 16) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMEnlistInternalTransaction(rpc ndr.Invoker, pUow msmqmq.XACTUOW) (PhIntXact msmqmp.RPC_INT_XACT_HANDLE, err error) {
	req := &r_QMEnlistInternalTransactionRequest{
		PUow: pUow,
	}
	var resp r_QMEnlistInternalTransactionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMEnlistInternalTransaction: %w", err)
		return
	}
	PhIntXact = resp.PhIntXact
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMEnlistInternalTransaction failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
