package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncCreatePrinterICRequest carries the [in] parameters of RpcAsyncCreatePrinterIC.
type rpcAsyncCreatePrinterICRequest struct {
	HPrinter          mspar.PRINTER_HANDLE
	PDevModeContainer mspar.DEVMODE_CONTAINER
}

func (*rpcAsyncCreatePrinterICRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncCreatePrinterIC
}

// rpcAsyncCreatePrinterICResponse carries the [out] parameters and return value of RpcAsyncCreatePrinterIC.
type rpcAsyncCreatePrinterICResponse struct {
	PHandle mspar.GDI_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcAsyncCreatePrinterIC calls RpcAsyncCreatePrinterIC (opnum 35) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncCreatePrinterIC(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pDevModeContainer mspar.DEVMODE_CONTAINER) (PHandle mspar.GDI_HANDLE, err error) {
	req := &rpcAsyncCreatePrinterICRequest{
		HPrinter:          hPrinter,
		PDevModeContainer: pDevModeContainer,
	}
	var resp rpcAsyncCreatePrinterICResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncCreatePrinterIC: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncCreatePrinterIC failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
