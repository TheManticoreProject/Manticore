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

// rpcSyncRegisterForRemoteNotificationsRequest carries the [in] parameters of RpcSyncRegisterForRemoteNotifications.
type rpcSyncRegisterForRemoteNotificationsRequest struct {
	HPrinter      mspar.PRINTER_HANDLE
	PNotifyFilter mspar.RpcPrintPropertiesCollection
}

func (*rpcSyncRegisterForRemoteNotificationsRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcSyncRegisterForRemoteNotifications
}

// rpcSyncRegisterForRemoteNotificationsResponse carries the [out] parameters and return value of RpcSyncRegisterForRemoteNotifications.
type rpcSyncRegisterForRemoteNotificationsResponse struct {
	PhRpcHandle mspar.RMTNTFY_HANDLE
	Status      ndr.DWORD `ndr:"retval"`
}

// RpcSyncRegisterForRemoteNotifications calls RpcSyncRegisterForRemoteNotifications (opnum 58) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcSyncRegisterForRemoteNotifications(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pNotifyFilter mspar.RpcPrintPropertiesCollection) (PhRpcHandle mspar.RMTNTFY_HANDLE, err error) {
	req := &rpcSyncRegisterForRemoteNotificationsRequest{
		HPrinter:      hPrinter,
		PNotifyFilter: pNotifyFilter,
	}
	var resp rpcSyncRegisterForRemoteNotificationsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSyncRegisterForRemoteNotifications: %w", err)
		return
	}
	PhRpcHandle = resp.PhRpcHandle
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcSyncRegisterForRemoteNotifications failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
