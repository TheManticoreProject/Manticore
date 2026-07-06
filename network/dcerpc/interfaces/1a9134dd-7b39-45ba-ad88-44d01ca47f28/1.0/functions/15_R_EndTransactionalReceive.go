package functions

// IDL source: [MS-MQMQ] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmq/56cc73e0-f57a-4bd9-aa45-861be5b85049
// A fetched copy is kept at ms-mqmq.idl in the interface directory.

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_EndTransactionalReceiveRequest carries the [in] parameters of R_EndTransactionalReceive.
type r_EndTransactionalReceiveRequest struct {
	PhContext   msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
	DwAck       ndr.DWORD
	DwRequestId ndr.DWORD
}

func (*r_EndTransactionalReceiveRequest) Opnum() uint16 {
	return RemoteRead.OpnumR_EndTransactionalReceive
}

// r_EndTransactionalReceiveResponse carries the [out] parameters and return value of R_EndTransactionalReceive.
type r_EndTransactionalReceiveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_EndTransactionalReceive calls R_EndTransactionalReceive (opnum 15) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_EndTransactionalReceive(rpc ndr.Invoker, phContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE, dwAck ndr.DWORD, dwRequestId ndr.DWORD) (err error) {
	req := &r_EndTransactionalReceiveRequest{
		PhContext:   phContext,
		DwAck:       dwAck,
		DwRequestId: dwRequestId,
	}
	var resp r_EndTransactionalReceiveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_EndTransactionalReceive: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_EndTransactionalReceive failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
