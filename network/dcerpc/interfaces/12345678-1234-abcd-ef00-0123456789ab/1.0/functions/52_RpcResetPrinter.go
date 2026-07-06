package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcResetPrinterRequest carries the [in] parameters of RpcResetPrinter.
type rpcResetPrinterRequest struct {
	HPrinter          msrprn.PRINTER_HANDLE
	PDatatype         *ndr.WSTR `ndr:"unique"`
	PDevModeContainer msrprn.DEVMODE_CONTAINER
}

func (*rpcResetPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcResetPrinter }

// rpcResetPrinterResponse carries the [out] parameters and return value of RpcResetPrinter.
type rpcResetPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcResetPrinter calls RpcResetPrinter (opnum 52) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcResetPrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pDatatype *ndr.WSTR, pDevModeContainer msrprn.DEVMODE_CONTAINER) (err error) {
	req := &rpcResetPrinterRequest{
		HPrinter:          hPrinter,
		PDatatype:         pDatatype,
		PDevModeContainer: pDevModeContainer,
	}
	var resp rpcResetPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcResetPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcResetPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
