package functions

import (
	"fmt"

	TermSrvNotification "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/11899a43-2b68-4a76-92e3-a3d6ad8c26ce/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcRegisterAsyncNotificationRequest carries the [in] parameters of RpcRegisterAsyncNotification.
type rpcRegisterAsyncNotificationRequest struct {
	SessionId int32
	Mask      mststs.TNotificationId
}

func (*rpcRegisterAsyncNotificationRequest) Opnum() uint16 {
	return TermSrvNotification.OpnumRpcRegisterAsyncNotification
}

// rpcRegisterAsyncNotificationResponse carries the [out] parameters and return value of RpcRegisterAsyncNotification.
type rpcRegisterAsyncNotificationResponse struct {
	PhNotify mststs.NOTIFY_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// RpcRegisterAsyncNotification calls RpcRegisterAsyncNotification (opnum 1) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcRegisterAsyncNotification(rpc ndr.Invoker, sessionId int32, mask mststs.TNotificationId) (PhNotify mststs.NOTIFY_HANDLE, err error) {
	req := &rpcRegisterAsyncNotificationRequest{
		SessionId: sessionId,
		Mask:      mask,
	}
	var resp rpcRegisterAsyncNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcRegisterAsyncNotification: %w", err)
		return
	}
	PhNotify = resp.PhNotify
	if uint32(resp.Status) != TermSrvNotification.StatusSuccess {
		err = fmt.Errorf("RpcRegisterAsyncNotification failed: %s", TermSrvNotification.StatusString(uint32(resp.Status)))
	}
	return
}
