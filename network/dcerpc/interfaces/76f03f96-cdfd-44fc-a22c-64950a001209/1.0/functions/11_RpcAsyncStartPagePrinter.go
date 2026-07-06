package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncStartPagePrinterRequest carries the [in] parameters of RpcAsyncStartPagePrinter.
type rpcAsyncStartPagePrinterRequest struct {
	HPrinter mspar.PRINTER_HANDLE
}

func (*rpcAsyncStartPagePrinterRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncStartPagePrinter
}

// rpcAsyncStartPagePrinterResponse carries the [out] parameters and return value of RpcAsyncStartPagePrinter.
type rpcAsyncStartPagePrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncStartPagePrinter calls RpcAsyncStartPagePrinter (opnum 11) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncStartPagePrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE) (err error) {
	req := &rpcAsyncStartPagePrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcAsyncStartPagePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncStartPagePrinter: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncStartPagePrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
