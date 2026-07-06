package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetQueueRequest carries the [in] parameters of FAX_SetQueue.
type fAX_SetQueueRequest struct {
	DwQueueStates ndr.DWORD
}

func (*fAX_SetQueueRequest) Opnum() uint16 { return fax.OpnumFAX_SetQueue }

// fAX_SetQueueResponse carries the [out] parameters and return value of FAX_SetQueue.
type fAX_SetQueueResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetQueue calls FAX_SetQueue (opnum 33) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetQueue(rpc ndr.Invoker, dwQueueStates ndr.DWORD) (err error) {
	req := &fAX_SetQueueRequest{
		DwQueueStates: dwQueueStates,
	}
	var resp fAX_SetQueueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetQueue: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetQueue failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
