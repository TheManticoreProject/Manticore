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

// rpcAsyncResetPrinterRequest carries the [in] parameters of RpcAsyncResetPrinter.
type rpcAsyncResetPrinterRequest struct {
	HPrinter          mspar.PRINTER_HANDLE
	PDatatype         *ndr.WSTR `ndr:"unique"`
	PDevModeContainer mspar.DEVMODE_CONTAINER
}

func (*rpcAsyncResetPrinterRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncResetPrinter }

// rpcAsyncResetPrinterResponse carries the [out] parameters and return value of RpcAsyncResetPrinter.
type rpcAsyncResetPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncResetPrinter calls RpcAsyncResetPrinter (opnum 69) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncResetPrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pDatatype *ndr.WSTR, pDevModeContainer mspar.DEVMODE_CONTAINER) (err error) {
	req := &rpcAsyncResetPrinterRequest{
		HPrinter:          hPrinter,
		PDatatype:         pDatatype,
		PDevModeContainer: pDevModeContainer,
	}
	var resp rpcAsyncResetPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncResetPrinter: %w", err)
		return
	}
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("RpcAsyncResetPrinter failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
