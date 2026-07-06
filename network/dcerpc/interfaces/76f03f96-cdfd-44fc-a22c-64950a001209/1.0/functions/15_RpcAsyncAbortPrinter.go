package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncAbortPrinterRequest carries the [in] parameters of RpcAsyncAbortPrinter.
type rpcAsyncAbortPrinterRequest struct {
	HPrinter mspar.PRINTER_HANDLE
}

func (*rpcAsyncAbortPrinterRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncAbortPrinter }

// rpcAsyncAbortPrinterResponse carries the [out] parameters and return value of RpcAsyncAbortPrinter.
type rpcAsyncAbortPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncAbortPrinter calls RpcAsyncAbortPrinter (opnum 15) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncAbortPrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE) (err error) {
	req := &rpcAsyncAbortPrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcAsyncAbortPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncAbortPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncAbortPrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
