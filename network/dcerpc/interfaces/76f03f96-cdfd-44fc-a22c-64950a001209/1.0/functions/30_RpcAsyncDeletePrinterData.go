package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
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
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("RpcAsyncDeletePrinterData failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
