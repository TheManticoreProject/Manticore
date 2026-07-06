package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncDeletePrinterRequest carries the [in] parameters of RpcAsyncDeletePrinter.
type rpcAsyncDeletePrinterRequest struct {
	HPrinter mspar.PRINTER_HANDLE
}

func (*rpcAsyncDeletePrinterRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePrinter
}

// rpcAsyncDeletePrinterResponse carries the [out] parameters and return value of RpcAsyncDeletePrinter.
type rpcAsyncDeletePrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePrinter calls RpcAsyncDeletePrinter (opnum 7) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE) (err error) {
	req := &rpcAsyncDeletePrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcAsyncDeletePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePrinter: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeletePrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
