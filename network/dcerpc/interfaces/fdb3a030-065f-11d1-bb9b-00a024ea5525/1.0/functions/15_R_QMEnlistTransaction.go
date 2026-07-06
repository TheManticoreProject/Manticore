package functions

// IDL source: [MS-MQMP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmp/a54c09de-1d72-47f0-9184-d7e5046b2ba1
// A fetched copy is kept at ms-mqmp.idl in the interface directory.

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// r_QMEnlistTransactionRequest carries the [in] parameters of R_QMEnlistTransaction.
type r_QMEnlistTransactionRequest struct {
	PUow     msmqmq.XACTUOW
	CbCookie ndr.DWORD
	PbCookie []uint8 `ndr:"ref,size_is=CbCookie"`
}

func (*r_QMEnlistTransactionRequest) Opnum() uint16 { return qmcomm.OpnumR_QMEnlistTransaction }

// r_QMEnlistTransactionResponse carries the [out] parameters and return value of R_QMEnlistTransaction.
type r_QMEnlistTransactionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_QMEnlistTransaction calls R_QMEnlistTransaction (opnum 15) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMEnlistTransaction(rpc ndr.Invoker, pUow msmqmq.XACTUOW, cbCookie ndr.DWORD, pbCookie []uint8) (err error) {
	req := &r_QMEnlistTransactionRequest{
		PUow:     pUow,
		CbCookie: cbCookie,
		PbCookie: pbCookie,
	}
	var resp r_QMEnlistTransactionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMEnlistTransaction: %w", err)
		return
	}
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMEnlistTransaction failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
