package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_CloseCursorRequest carries the [in] parameters of R_CloseCursor.
type r_CloseCursorRequest struct {
	PhContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
	HCursor   ndr.DWORD
}

func (*r_CloseCursorRequest) Opnum() uint16 { return RemoteRead.OpnumR_CloseCursor }

// r_CloseCursorResponse carries the [out] parameters and return value of R_CloseCursor.
type r_CloseCursorResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_CloseCursor calls R_CloseCursor (opnum 5) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_CloseCursor(rpc ndr.Invoker, phContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE, hCursor ndr.DWORD) (err error) {
	req := &r_CloseCursorRequest{
		PhContext: phContext,
		HCursor:   hCursor,
	}
	var resp r_CloseCursorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_CloseCursor: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_CloseCursor failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
