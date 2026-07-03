package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcRouterRefreshPrinterChangeNotificationRequest carries the [in] parameters of RpcRouterRefreshPrinterChangeNotification.
type rpcRouterRefreshPrinterChangeNotificationRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	DwColor  ndr.DWORD
	POptions *msrprn.RPC_V2_NOTIFY_OPTIONS `ndr:"unique"`
}

func (*rpcRouterRefreshPrinterChangeNotificationRequest) Opnum() uint16 {
	return winspool.OpnumRpcRouterRefreshPrinterChangeNotification
}

// rpcRouterRefreshPrinterChangeNotificationResponse carries the [out] parameters and return value of RpcRouterRefreshPrinterChangeNotification.
type rpcRouterRefreshPrinterChangeNotificationResponse struct {
	PpInfo *msrprn.RPC_V2_NOTIFY_INFO `ndr:"unique"`
	Status ndr.DWORD                  `ndr:"retval"`
}

// RpcRouterRefreshPrinterChangeNotification calls RpcRouterRefreshPrinterChangeNotification (opnum 67) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcRouterRefreshPrinterChangeNotification(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, dwColor ndr.DWORD, pOptions *msrprn.RPC_V2_NOTIFY_OPTIONS) (PpInfo *msrprn.RPC_V2_NOTIFY_INFO, err error) {
	req := &rpcRouterRefreshPrinterChangeNotificationRequest{
		HPrinter: hPrinter,
		DwColor:  dwColor,
		POptions: pOptions,
	}
	var resp rpcRouterRefreshPrinterChangeNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcRouterRefreshPrinterChangeNotification: %w", err)
		return
	}
	PpInfo = resp.PpInfo
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcRouterRefreshPrinterChangeNotification failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
