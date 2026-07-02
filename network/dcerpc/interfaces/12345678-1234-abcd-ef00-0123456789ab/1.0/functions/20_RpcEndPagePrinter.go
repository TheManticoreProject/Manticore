package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEndPagePrinterRequest carries the [in] parameters of RpcEndPagePrinter.
type rpcEndPagePrinterRequest struct {
	HPrinter structures.PRINTER_HANDLE
}

func (*rpcEndPagePrinterRequest) Opnum() uint16 { return winspool.OpnumRpcEndPagePrinter }

// rpcEndPagePrinterResponse carries the [out] parameters and return value of RpcEndPagePrinter.
type rpcEndPagePrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcEndPagePrinter calls RpcEndPagePrinter (opnum 20) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEndPagePrinter(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE) (err error) {
	req := &rpcEndPagePrinterRequest{
		HPrinter: hPrinter,
	}
	var resp rpcEndPagePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEndPagePrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEndPagePrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
