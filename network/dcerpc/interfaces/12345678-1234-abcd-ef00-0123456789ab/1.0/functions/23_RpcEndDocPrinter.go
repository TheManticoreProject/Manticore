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

// rpcEndDocPrinterRequest carries the [in] parameters of RpcEndDocPrinter.
type rpcEndDocPrinterRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
}

func (*rpcEndDocPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcEndDocPrinter }

// rpcEndDocPrinterResponse carries the [out] parameters and return value of RpcEndDocPrinter.
type rpcEndDocPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcEndDocPrinter calls RpcEndDocPrinter (opnum 23) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEndDocPrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE) (err error) {
	req := &rpcEndDocPrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcEndDocPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEndDocPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEndDocPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
