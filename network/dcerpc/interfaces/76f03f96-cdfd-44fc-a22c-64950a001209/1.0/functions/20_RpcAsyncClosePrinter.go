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

// rpcAsyncClosePrinterRequest carries the [in] parameters of RpcAsyncClosePrinter.
type rpcAsyncClosePrinterRequest struct {
	PhPrinter mspar.PRINTER_HANDLE
}

func (*rpcAsyncClosePrinterRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncClosePrinter }

// rpcAsyncClosePrinterResponse carries the [out] parameters and return value of RpcAsyncClosePrinter.
type rpcAsyncClosePrinterResponse struct {
	PhPrinter mspar.PRINTER_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncClosePrinter calls RpcAsyncClosePrinter (opnum 20) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncClosePrinter(rpc ndr.Invoker, phPrinter mspar.PRINTER_HANDLE) (PhPrinter mspar.PRINTER_HANDLE, err error) {
	req := &rpcAsyncClosePrinterRequest{
		PhPrinter: phPrinter,
	}
	var resp rpcAsyncClosePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncClosePrinter: %w", err)
		return
	}
	PhPrinter = resp.PhPrinter
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncClosePrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
