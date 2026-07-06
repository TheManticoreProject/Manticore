package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcSyncUnRegisterForRemoteNotificationsRequest carries the [in] parameters of RpcSyncUnRegisterForRemoteNotifications.
type rpcSyncUnRegisterForRemoteNotificationsRequest struct {
	PhRpcHandle mspar.RMTNTFY_HANDLE
}

func (*rpcSyncUnRegisterForRemoteNotificationsRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcSyncUnRegisterForRemoteNotifications
}

// rpcSyncUnRegisterForRemoteNotificationsResponse carries the [out] parameters and return value of RpcSyncUnRegisterForRemoteNotifications.
type rpcSyncUnRegisterForRemoteNotificationsResponse struct {
	PhRpcHandle mspar.RMTNTFY_HANDLE
	Status      ndr.DWORD `ndr:"retval"`
}

// RpcSyncUnRegisterForRemoteNotifications calls RpcSyncUnRegisterForRemoteNotifications (opnum 59) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcSyncUnRegisterForRemoteNotifications(rpc ndr.Invoker, phRpcHandle mspar.RMTNTFY_HANDLE) (PhRpcHandle mspar.RMTNTFY_HANDLE, err error) {
	req := &rpcSyncUnRegisterForRemoteNotificationsRequest{
		PhRpcHandle: phRpcHandle,
	}
	var resp rpcSyncUnRegisterForRemoteNotificationsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSyncUnRegisterForRemoteNotifications: %w", err)
		return
	}
	PhRpcHandle = resp.PhRpcHandle
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcSyncUnRegisterForRemoteNotifications failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
