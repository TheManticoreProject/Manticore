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

// rpcAsyncStartDocPrinterRequest carries the [in] parameters of RpcAsyncStartDocPrinter.
type rpcAsyncStartDocPrinterRequest struct {
	HPrinter          mspar.PRINTER_HANDLE
	PDocInfoContainer mspar.DOC_INFO_CONTAINER
}

func (*rpcAsyncStartDocPrinterRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncStartDocPrinter
}

// rpcAsyncStartDocPrinterResponse carries the [out] parameters and return value of RpcAsyncStartDocPrinter.
type rpcAsyncStartDocPrinterResponse struct {
	PJobId ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncStartDocPrinter calls RpcAsyncStartDocPrinter (opnum 10) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncStartDocPrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pDocInfoContainer mspar.DOC_INFO_CONTAINER) (PJobId ndr.DWORD, err error) {
	req := &rpcAsyncStartDocPrinterRequest{
		HPrinter:          hPrinter,
		PDocInfoContainer: pDocInfoContainer,
	}
	var resp rpcAsyncStartDocPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncStartDocPrinter: %w", err)
		return
	}
	PJobId = resp.PJobId
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("RpcAsyncStartDocPrinter failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
