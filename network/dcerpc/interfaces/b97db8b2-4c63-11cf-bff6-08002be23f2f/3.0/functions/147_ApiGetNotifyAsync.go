package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetNotifyAsyncRequest carries the [in] parameters of ApiGetNotifyAsync.
type apiGetNotifyAsyncRequest struct {
	HNotify mscmrp.HNOTIFY_RPC
}

func (*apiGetNotifyAsyncRequest) Opnum() uint16 { return clusapi.OpnumApiGetNotifyAsync }

// apiGetNotifyAsyncResponse carries the [out] parameters and return value of ApiGetNotifyAsync.
type apiGetNotifyAsyncResponse struct {
	Notifications      []*mscmrp.NOTIFICATION_DATA_ASYNC_RPC `ndr:"elem=unique,ref,size_is=DwNumNotifications"`
	DwNumNotifications ndr.DWORD
	Status             ndr.DWORD `ndr:"retval"`
}

// ApiGetNotifyAsync calls ApiGetNotifyAsync (opnum 147) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNotifyAsync(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC) (Notifications []*mscmrp.NOTIFICATION_DATA_ASYNC_RPC, DwNumNotifications ndr.DWORD, err error) {
	req := &apiGetNotifyAsyncRequest{
		HNotify: hNotify,
	}
	var resp apiGetNotifyAsyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNotifyAsync: %w", err)
		return
	}
	Notifications = resp.Notifications
	DwNumNotifications = resp.DwNumNotifications
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNotifyAsync failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
