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
)

// r_QMCommitTransactionRequest carries the [in] parameters of R_QMCommitTransaction.
type r_QMCommitTransactionRequest struct {
	PhIntXact msmqmp.RPC_INT_XACT_HANDLE
}

func (*r_QMCommitTransactionRequest) Opnum() uint16 { return qmcomm.OpnumR_QMCommitTransaction }

// r_QMCommitTransactionResponse carries the [out] parameters and return value of R_QMCommitTransaction.
type r_QMCommitTransactionResponse struct {
	PhIntXact msmqmp.RPC_INT_XACT_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// R_QMCommitTransaction calls R_QMCommitTransaction (opnum 17) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMCommitTransaction(rpc ndr.Invoker, phIntXact msmqmp.RPC_INT_XACT_HANDLE) (PhIntXact msmqmp.RPC_INT_XACT_HANDLE, err error) {
	req := &r_QMCommitTransactionRequest{
		PhIntXact: phIntXact,
	}
	var resp r_QMCommitTransactionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMCommitTransaction: %w", err)
		return
	}
	PhIntXact = resp.PhIntXact
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMCommitTransaction failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
