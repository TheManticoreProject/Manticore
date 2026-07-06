package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncDeletePrinterDataRequest carries the [in] parameters of RpcAsyncDeletePrinterData.
type rpcAsyncDeletePrinterDataRequest struct {
	HPrinter   mspar.PRINTER_HANDLE
	PValueName ndr.WSTR
}

func (*rpcAsyncDeletePrinterDataRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePrinterData
}

// rpcAsyncDeletePrinterDataResponse carries the [out] parameters and return value of RpcAsyncDeletePrinterData.
type rpcAsyncDeletePrinterDataResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePrinterData calls RpcAsyncDeletePrinterData (opnum 30) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePrinterData(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pValueName ndr.WSTR) (err error) {
	req := &rpcAsyncDeletePrinterDataRequest{
		HPrinter:   hPrinter,
		PValueName: pValueName,
	}
	var resp rpcAsyncDeletePrinterDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePrinterData: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeletePrinterData failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
