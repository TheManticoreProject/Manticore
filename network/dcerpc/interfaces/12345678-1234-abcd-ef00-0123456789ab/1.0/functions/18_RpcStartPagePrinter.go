package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcStartPagePrinterRequest carries the [in] parameters of RpcStartPagePrinter.
type rpcStartPagePrinterRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
}

func (*rpcStartPagePrinterRequest) Opnum() uint16 { return winspool.OpnumRpcStartPagePrinter }

// rpcStartPagePrinterResponse carries the [out] parameters and return value of RpcStartPagePrinter.
type rpcStartPagePrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcStartPagePrinter calls RpcStartPagePrinter (opnum 18) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcStartPagePrinter(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE) (err error) {
	req := &rpcStartPagePrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcStartPagePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcStartPagePrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcStartPagePrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
