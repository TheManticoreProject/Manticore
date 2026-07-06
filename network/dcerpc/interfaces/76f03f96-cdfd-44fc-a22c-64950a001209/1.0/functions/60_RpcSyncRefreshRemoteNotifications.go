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

// rpcSyncRefreshRemoteNotificationsRequest carries the [in] parameters of RpcSyncRefreshRemoteNotifications.
type rpcSyncRefreshRemoteNotificationsRequest struct {
	HRpcHandle    mspar.RMTNTFY_HANDLE
	PNotifyFilter mspar.RpcPrintPropertiesCollection
}

func (*rpcSyncRefreshRemoteNotificationsRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcSyncRefreshRemoteNotifications
}

// rpcSyncRefreshRemoteNotificationsResponse carries the [out] parameters and return value of RpcSyncRefreshRemoteNotifications.
type rpcSyncRefreshRemoteNotificationsResponse struct {
	PpNotifyData mspar.RpcPrintPropertiesCollection
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcSyncRefreshRemoteNotifications calls RpcSyncRefreshRemoteNotifications (opnum 60) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcSyncRefreshRemoteNotifications(rpc ndr.Invoker, hRpcHandle mspar.RMTNTFY_HANDLE, pNotifyFilter mspar.RpcPrintPropertiesCollection) (PpNotifyData mspar.RpcPrintPropertiesCollection, err error) {
	req := &rpcSyncRefreshRemoteNotificationsRequest{
		HRpcHandle:    hRpcHandle,
		PNotifyFilter: pNotifyFilter,
	}
	var resp rpcSyncRefreshRemoteNotificationsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSyncRefreshRemoteNotifications: %w", err)
		return
	}
	PpNotifyData = resp.PpNotifyData
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcSyncRefreshRemoteNotifications failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
