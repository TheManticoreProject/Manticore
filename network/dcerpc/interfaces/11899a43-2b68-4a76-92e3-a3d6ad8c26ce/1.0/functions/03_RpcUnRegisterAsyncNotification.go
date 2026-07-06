package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	TermSrvNotification "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/11899a43-2b68-4a76-92e3-a3d6ad8c26ce/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcUnRegisterAsyncNotificationRequest carries the [in] parameters of RpcUnRegisterAsyncNotification.
type rpcUnRegisterAsyncNotificationRequest struct {
	PhNotify mststs.NOTIFY_HANDLE
}

func (*rpcUnRegisterAsyncNotificationRequest) Opnum() uint16 {
	return TermSrvNotification.OpnumRpcUnRegisterAsyncNotification
}

// rpcUnRegisterAsyncNotificationResponse carries the [out] parameters and return value of RpcUnRegisterAsyncNotification.
type rpcUnRegisterAsyncNotificationResponse struct {
	PhNotify mststs.NOTIFY_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// RpcUnRegisterAsyncNotification calls RpcUnRegisterAsyncNotification (opnum 3) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcUnRegisterAsyncNotification(rpc ndr.Invoker, phNotify mststs.NOTIFY_HANDLE) (PhNotify mststs.NOTIFY_HANDLE, err error) {
	req := &rpcUnRegisterAsyncNotificationRequest{
		PhNotify: phNotify,
	}
	var resp rpcUnRegisterAsyncNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcUnRegisterAsyncNotification: %w", err)
		return
	}
	PhNotify = resp.PhNotify
	if uint32(resp.Status) != TermSrvNotification.StatusSuccess {
		err = fmt.Errorf("RpcUnRegisterAsyncNotification failed: %s", TermSrvNotification.StatusString(uint32(resp.Status)))
	}
	return
}
