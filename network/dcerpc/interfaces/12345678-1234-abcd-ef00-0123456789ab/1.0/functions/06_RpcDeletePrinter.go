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

// rpcDeletePrinterRequest carries the [in] parameters of RpcDeletePrinter.
type rpcDeletePrinterRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
}

func (*rpcDeletePrinterRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrinter }

// rpcDeletePrinterResponse carries the [out] parameters and return value of RpcDeletePrinter.
type rpcDeletePrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrinter calls RpcDeletePrinter (opnum 6) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE) (err error) {
	req := &rpcDeletePrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcDeletePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
