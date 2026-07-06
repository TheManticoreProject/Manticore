package functions

// IDL source: [MS-MQMP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmp/a54c09de-1d72-47f0-9184-d7e5046b2ba1
// A fetched copy is kept at ms-mqmp.idl in the interface directory.

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_QMGetRemoteQueueNameRequest carries the [in] parameters of R_QMGetRemoteQueueName.
type r_QMGetRemoteQueueNameRequest struct {
	PQueue              ndr.DWORD
	LplpRemoteQueueName *ndr.WSTR `ndr:"ptr"`
}

func (*r_QMGetRemoteQueueNameRequest) Opnum() uint16 { return qmcomm.OpnumR_QMGetRemoteQueueName }

// r_QMGetRemoteQueueNameResponse carries the [out] parameters and return value of R_QMGetRemoteQueueName.
type r_QMGetRemoteQueueNameResponse struct {
	LplpRemoteQueueName *ndr.WSTR `ndr:"ptr"`
	Status              ndr.DWORD `ndr:"retval"`
}

// R_QMGetRemoteQueueName calls R_QMGetRemoteQueueName (opnum 1) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMGetRemoteQueueName(rpc ndr.Invoker, pQueue ndr.DWORD, lplpRemoteQueueName *ndr.WSTR) (LplpRemoteQueueName *ndr.WSTR, err error) {
	req := &r_QMGetRemoteQueueNameRequest{
		PQueue:              pQueue,
		LplpRemoteQueueName: lplpRemoteQueueName,
	}
	var resp r_QMGetRemoteQueueNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMGetRemoteQueueName: %w", err)
		return
	}
	LplpRemoteQueueName = resp.LplpRemoteQueueName
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMGetRemoteQueueName failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
