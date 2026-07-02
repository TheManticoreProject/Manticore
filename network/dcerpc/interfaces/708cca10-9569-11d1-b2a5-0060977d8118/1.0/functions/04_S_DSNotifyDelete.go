package functions

import (
	"fmt"

	dscomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/708cca10-9569-11d1-b2a5-0060977d8118/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSNotifyDeleteRequest carries the [in] parameters of S_DSNotifyDelete.
type s_DSNotifyDeleteRequest struct {
	Handle msmqds.PCONTEXT_HANDLE_DELETE_TYPE
}

func (*s_DSNotifyDeleteRequest) Opnum() uint16 { return dscomm2.OpnumS_DSNotifyDelete }

// s_DSNotifyDeleteResponse carries the [out] parameters and return value of S_DSNotifyDelete.
type s_DSNotifyDeleteResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// S_DSNotifyDelete calls S_DSNotifyDelete (opnum 4) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSNotifyDelete(rpc ndr.Invoker, handle msmqds.PCONTEXT_HANDLE_DELETE_TYPE) (err error) {
	req := &s_DSNotifyDeleteRequest{
		Handle: handle,
	}
	var resp s_DSNotifyDeleteResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSNotifyDelete: %w", err)
		return
	}
	if uint32(resp.Status) != dscomm2.StatusSuccess {
		err = fmt.Errorf("S_DSNotifyDelete failed: %s", dscomm2.StatusString(uint32(resp.Status)))
	}
	return
}
