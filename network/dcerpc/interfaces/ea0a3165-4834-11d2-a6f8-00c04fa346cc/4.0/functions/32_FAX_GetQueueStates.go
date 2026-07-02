package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetQueueStatesRequest carries the [in] parameters of FAX_GetQueueStates.
type fAX_GetQueueStatesRequest struct {
}

func (*fAX_GetQueueStatesRequest) Opnum() uint16 { return fax.OpnumFAX_GetQueueStates }

// fAX_GetQueueStatesResponse carries the [out] parameters and return value of FAX_GetQueueStates.
type fAX_GetQueueStatesResponse struct {
	PdwQueueStates ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// FAX_GetQueueStates calls FAX_GetQueueStates (opnum 32) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetQueueStates(rpc ndr.Invoker) (PdwQueueStates ndr.DWORD, err error) {
	req := &fAX_GetQueueStatesRequest{}
	var resp fAX_GetQueueStatesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetQueueStates: %w", err)
		return
	}
	PdwQueueStates = resp.PdwQueueStates
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetQueueStates failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
