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

// rpcCreatePrinterICRequest carries the [in] parameters of RpcCreatePrinterIC.
type rpcCreatePrinterICRequest struct {
	HPrinter          msrprn.PRINTER_HANDLE
	PDevModeContainer msrprn.DEVMODE_CONTAINER
}

func (*rpcCreatePrinterICRequest) Opnum() uint16 { return winspool.OpnumRpcCreatePrinterIC }

// rpcCreatePrinterICResponse carries the [out] parameters and return value of RpcCreatePrinterIC.
type rpcCreatePrinterICResponse struct {
	PHandle msrprn.GDI_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcCreatePrinterIC calls RpcCreatePrinterIC (opnum 40) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcCreatePrinterIC(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pDevModeContainer msrprn.DEVMODE_CONTAINER) (PHandle msrprn.GDI_HANDLE, err error) {
	req := &rpcCreatePrinterICRequest{
		HPrinter:          hPrinter,
		PDevModeContainer: pDevModeContainer,
	}
	var resp rpcCreatePrinterICResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcCreatePrinterIC: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcCreatePrinterIC failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
