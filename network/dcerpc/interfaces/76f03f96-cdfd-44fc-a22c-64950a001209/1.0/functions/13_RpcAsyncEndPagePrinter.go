package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncEndPagePrinterRequest carries the [in] parameters of RpcAsyncEndPagePrinter.
type rpcAsyncEndPagePrinterRequest struct {
	HPrinter mspar.PRINTER_HANDLE
}

func (*rpcAsyncEndPagePrinterRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEndPagePrinter
}

// rpcAsyncEndPagePrinterResponse carries the [out] parameters and return value of RpcAsyncEndPagePrinter.
type rpcAsyncEndPagePrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEndPagePrinter calls RpcAsyncEndPagePrinter (opnum 13) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEndPagePrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE) (err error) {
	req := &rpcAsyncEndPagePrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcAsyncEndPagePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEndPagePrinter: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEndPagePrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
