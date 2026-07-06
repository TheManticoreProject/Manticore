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

// r_CancelReceiveRequest carries the [in] parameters of R_CancelReceive.
type r_CancelReceiveRequest struct {
	PhContext   msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
	DwRequestId ndr.DWORD
}

func (*r_CancelReceiveRequest) Opnum() uint16 { return RemoteRead.OpnumR_CancelReceive }

// r_CancelReceiveResponse carries the [out] parameters and return value of R_CancelReceive.
type r_CancelReceiveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_CancelReceive calls R_CancelReceive (opnum 8) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_CancelReceive(rpc ndr.Invoker, phContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE, dwRequestId ndr.DWORD) (err error) {
	req := &r_CancelReceiveRequest{
		PhContext:   phContext,
		DwRequestId: dwRequestId,
	}
	var resp r_CancelReceiveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_CancelReceive: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_CancelReceive failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
