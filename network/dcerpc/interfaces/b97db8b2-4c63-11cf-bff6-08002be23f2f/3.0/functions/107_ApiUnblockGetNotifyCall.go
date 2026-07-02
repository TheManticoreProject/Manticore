package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiUnblockGetNotifyCallRequest carries the [in] parameters of ApiUnblockGetNotifyCall.
type apiUnblockGetNotifyCallRequest struct {
	HNotify mscmrp.HNOTIFY_RPC
}

func (*apiUnblockGetNotifyCallRequest) Opnum() uint16 { return clusapi.OpnumApiUnblockGetNotifyCall }

// apiUnblockGetNotifyCallResponse carries the [out] parameters and return value of ApiUnblockGetNotifyCall.
type apiUnblockGetNotifyCallResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// ApiUnblockGetNotifyCall calls ApiUnblockGetNotifyCall (opnum 107) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiUnblockGetNotifyCall(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC) (err error) {
	req := &apiUnblockGetNotifyCallRequest{
		HNotify: hNotify,
	}
	var resp apiUnblockGetNotifyCallResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiUnblockGetNotifyCall: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiUnblockGetNotifyCall failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
