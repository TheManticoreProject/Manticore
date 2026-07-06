package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncEndDocPrinterRequest carries the [in] parameters of RpcAsyncEndDocPrinter.
type rpcAsyncEndDocPrinterRequest struct {
	HPrinter mspar.PRINTER_HANDLE
}

func (*rpcAsyncEndDocPrinterRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEndDocPrinter
}

// rpcAsyncEndDocPrinterResponse carries the [out] parameters and return value of RpcAsyncEndDocPrinter.
type rpcAsyncEndDocPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEndDocPrinter calls RpcAsyncEndDocPrinter (opnum 14) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEndDocPrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE) (err error) {
	req := &rpcAsyncEndDocPrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcAsyncEndDocPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEndDocPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEndDocPrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
