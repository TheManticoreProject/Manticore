package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_EndReceiveRequest carries the [in] parameters of R_EndReceive.
type r_EndReceiveRequest struct {
	PhContext   msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
	DwAck       ndr.DWORD
	DwRequestId ndr.DWORD
}

func (*r_EndReceiveRequest) Opnum() uint16 { return RemoteRead.OpnumR_EndReceive }

// r_EndReceiveResponse carries the [out] parameters and return value of R_EndReceive.
type r_EndReceiveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_EndReceive calls R_EndReceive (opnum 9) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_EndReceive(rpc ndr.Invoker, phContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE, dwAck ndr.DWORD, dwRequestId ndr.DWORD) (err error) {
	req := &r_EndReceiveRequest{
		PhContext:   phContext,
		DwAck:       dwAck,
		DwRequestId: dwRequestId,
	}
	var resp r_EndReceiveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_EndReceive: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_EndReceive failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
