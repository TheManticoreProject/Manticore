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

// rpcAsyncGetRemoteNotificationsRequest carries the [in] parameters of RpcAsyncGetRemoteNotifications.
type rpcAsyncGetRemoteNotificationsRequest struct {
	HRpcHandle mspar.RMTNTFY_HANDLE
}

func (*rpcAsyncGetRemoteNotificationsRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetRemoteNotifications
}

// rpcAsyncGetRemoteNotificationsResponse carries the [out] parameters and return value of RpcAsyncGetRemoteNotifications.
type rpcAsyncGetRemoteNotificationsResponse struct {
	PpNotifyData mspar.RpcPrintPropertiesCollection
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetRemoteNotifications calls RpcAsyncGetRemoteNotifications (opnum 61) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetRemoteNotifications(rpc ndr.Invoker, hRpcHandle mspar.RMTNTFY_HANDLE) (PpNotifyData mspar.RpcPrintPropertiesCollection, err error) {
	req := &rpcAsyncGetRemoteNotificationsRequest{
		HRpcHandle: hRpcHandle,
	}
	var resp rpcAsyncGetRemoteNotificationsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetRemoteNotifications: %w", err)
		return
	}
	PpNotifyData = resp.PpNotifyData
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetRemoteNotifications failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
