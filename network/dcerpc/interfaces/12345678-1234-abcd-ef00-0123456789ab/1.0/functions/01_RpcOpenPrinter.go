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

// rpcOpenPrinterRequest carries the [in] parameters of RpcOpenPrinter.
type rpcOpenPrinterRequest struct {
	PPrinterName      *ndr.WSTR `ndr:"unique"`
	PDatatype         *ndr.WSTR `ndr:"unique"`
	PDevModeContainer msrprn.DEVMODE_CONTAINER
	AccessRequired    ndr.DWORD
}

func (*rpcOpenPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcOpenPrinter }

// rpcOpenPrinterResponse carries the [out] parameters and return value of RpcOpenPrinter.
type rpcOpenPrinterResponse struct {
	PHandle msrprn.PRINTER_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcOpenPrinter calls RpcOpenPrinter (opnum 1) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcOpenPrinter(rpc ndr.Invoker, pPrinterName *ndr.WSTR, pDatatype *ndr.WSTR, pDevModeContainer msrprn.DEVMODE_CONTAINER, accessRequired ndr.DWORD) (PHandle msrprn.PRINTER_HANDLE, err error) {
	req := &rpcOpenPrinterRequest{
		PPrinterName:      pPrinterName,
		PDatatype:         pDatatype,
		PDevModeContainer: pDevModeContainer,
		AccessRequired:    accessRequired,
	}
	var resp rpcOpenPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcOpenPrinter: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcOpenPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
