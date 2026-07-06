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

// rpcSetPrinterRequest carries the [in] parameters of RpcSetPrinter.
type rpcSetPrinterRequest struct {
	HPrinter           msrprn.PRINTER_HANDLE
	PPrinterContainer  msrprn.PRINTER_CONTAINER
	PDevModeContainer  msrprn.DEVMODE_CONTAINER
	PSecurityContainer msrprn.SECURITY_CONTAINER
	Command            ndr.DWORD
}

func (*rpcSetPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcSetPrinter }

// rpcSetPrinterResponse carries the [out] parameters and return value of RpcSetPrinter.
type rpcSetPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcSetPrinter calls RpcSetPrinter (opnum 7) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSetPrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pPrinterContainer msrprn.PRINTER_CONTAINER, pDevModeContainer msrprn.DEVMODE_CONTAINER, pSecurityContainer msrprn.SECURITY_CONTAINER, command ndr.DWORD) (err error) {
	req := &rpcSetPrinterRequest{
		HPrinter:           hPrinter,
		PPrinterContainer:  pPrinterContainer,
		PDevModeContainer:  pDevModeContainer,
		PSecurityContainer: pSecurityContainer,
		Command:            command,
	}
	var resp rpcSetPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSetPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSetPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
