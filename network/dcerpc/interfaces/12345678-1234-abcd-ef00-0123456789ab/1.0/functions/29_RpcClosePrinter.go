package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcClosePrinterRequest carries the [in] parameters of RpcClosePrinter.
type rpcClosePrinterRequest struct {
	PhPrinter msrprn.PRINTER_HANDLE
}

func (*rpcClosePrinterRequest) Opnum() uint16 { return winspool.OpnumRpcClosePrinter }

// rpcClosePrinterResponse carries the [out] parameters and return value of RpcClosePrinter.
type rpcClosePrinterResponse struct {
	PhPrinter msrprn.PRINTER_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcClosePrinter calls RpcClosePrinter (opnum 29) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcClosePrinter(rpc ndr.Invoker, phPrinter msrprn.PRINTER_HANDLE) (PhPrinter msrprn.PRINTER_HANDLE, err error) {
	req := &rpcClosePrinterRequest{
		PhPrinter: phPrinter,
	}
	var resp rpcClosePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcClosePrinter: %w", err)
		return
	}
	PhPrinter = resp.PhPrinter
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcClosePrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
