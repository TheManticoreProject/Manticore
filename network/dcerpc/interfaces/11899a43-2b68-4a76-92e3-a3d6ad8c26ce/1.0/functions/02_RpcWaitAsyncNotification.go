package functions

import (
	"fmt"

	TermSrvNotification "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/11899a43-2b68-4a76-92e3-a3d6ad8c26ce/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWaitAsyncNotificationRequest carries the [in] parameters of RpcWaitAsyncNotification.
type rpcWaitAsyncNotificationRequest struct {
	HNotify mststs.NOTIFY_HANDLE
}

func (*rpcWaitAsyncNotificationRequest) Opnum() uint16 {
	return TermSrvNotification.OpnumRpcWaitAsyncNotification
}

// rpcWaitAsyncNotificationResponse carries the [out] parameters and return value of RpcWaitAsyncNotification.
type rpcWaitAsyncNotificationResponse struct {
	SessionChange []mststs.SESSION_CHANGE `ndr:"unique,conformant"`
	PEntries      ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcWaitAsyncNotification calls RpcWaitAsyncNotification (opnum 2) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWaitAsyncNotification(rpc ndr.Invoker, hNotify mststs.NOTIFY_HANDLE) (SessionChange []mststs.SESSION_CHANGE, PEntries ndr.DWORD, err error) {
	req := &rpcWaitAsyncNotificationRequest{
		HNotify: hNotify,
	}
	var resp rpcWaitAsyncNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWaitAsyncNotification: %w", err)
		return
	}
	SessionChange = resp.SessionChange
	PEntries = resp.PEntries
	if uint32(resp.Status) != TermSrvNotification.StatusSuccess {
		err = fmt.Errorf("RpcWaitAsyncNotification failed: %s", TermSrvNotification.StatusString(uint32(resp.Status)))
	}
	return
}
