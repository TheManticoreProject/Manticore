package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqrr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqrr"
)

// r_CreateCursorRequest carries the [in] parameters of R_CreateCursor.
type r_CreateCursorRequest struct {
	PhContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE
}

func (*r_CreateCursorRequest) Opnum() uint16 { return RemoteRead.OpnumR_CreateCursor }

// r_CreateCursorResponse carries the [out] parameters and return value of R_CreateCursor.
type r_CreateCursorResponse struct {
	PhCursor ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// R_CreateCursor calls R_CreateCursor (opnum 4) ([MS-MQRR] — verify the parameter
// modeling and status handling).
func R_CreateCursor(rpc ndr.Invoker, phContext msmqrr.QUEUE_CONTEXT_HANDLE_NOSERIALIZE) (PhCursor ndr.DWORD, err error) {
	req := &r_CreateCursorRequest{
		PhContext: phContext,
	}
	var resp r_CreateCursorResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_CreateCursor: %w", err)
		return
	}
	PhCursor = resp.PhCursor
	if uint32(resp.Status) != RemoteRead.StatusSuccess {
		err = fmt.Errorf("R_CreateCursor failed: %s", RemoteRead.StatusString(uint32(resp.Status)))
	}
	return
}
