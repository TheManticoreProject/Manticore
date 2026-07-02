package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcFindClosePrinterChangeNotificationRequest carries the [in] parameters of RpcFindClosePrinterChangeNotification.
type rpcFindClosePrinterChangeNotificationRequest struct {
	HPrinter structures.PRINTER_HANDLE
}

func (*rpcFindClosePrinterChangeNotificationRequest) Opnum() uint16 {
	return winspool.OpnumRpcFindClosePrinterChangeNotification
}

// rpcFindClosePrinterChangeNotificationResponse carries the [out] parameters and return value of RpcFindClosePrinterChangeNotification.
type rpcFindClosePrinterChangeNotificationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcFindClosePrinterChangeNotification calls RpcFindClosePrinterChangeNotification (opnum 56) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcFindClosePrinterChangeNotification(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE) (err error) {
	req := &rpcFindClosePrinterChangeNotificationRequest{
		HPrinter: hPrinter,
	}
	var resp rpcFindClosePrinterChangeNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcFindClosePrinterChangeNotification: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcFindClosePrinterChangeNotification failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
