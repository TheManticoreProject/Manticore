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

// rpcAsyncSetPrinterRequest carries the [in] parameters of RpcAsyncSetPrinter.
type rpcAsyncSetPrinterRequest struct {
	HPrinter           mspar.PRINTER_HANDLE
	PPrinterContainer  mspar.PRINTER_CONTAINER
	PDevModeContainer  mspar.DEVMODE_CONTAINER
	PSecurityContainer mspar.SECURITY_CONTAINER
	Command            ndr.DWORD
}

func (*rpcAsyncSetPrinterRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncSetPrinter }

// rpcAsyncSetPrinterResponse carries the [out] parameters and return value of RpcAsyncSetPrinter.
type rpcAsyncSetPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncSetPrinter calls RpcAsyncSetPrinter (opnum 8) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncSetPrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pPrinterContainer mspar.PRINTER_CONTAINER, pDevModeContainer mspar.DEVMODE_CONTAINER, pSecurityContainer mspar.SECURITY_CONTAINER, command ndr.DWORD) (err error) {
	req := &rpcAsyncSetPrinterRequest{
		HPrinter:           hPrinter,
		PPrinterContainer:  pPrinterContainer,
		PDevModeContainer:  pDevModeContainer,
		PSecurityContainer: pSecurityContainer,
		Command:            command,
	}
	var resp rpcAsyncSetPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncSetPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncSetPrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
