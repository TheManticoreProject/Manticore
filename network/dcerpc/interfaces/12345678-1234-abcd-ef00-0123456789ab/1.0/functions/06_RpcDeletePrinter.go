package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcDeletePrinterRequest carries the [in] parameters of RpcDeletePrinter.
type rpcDeletePrinterRequest struct {
	HPrinter structures.PRINTER_HANDLE
}

func (*rpcDeletePrinterRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrinter }

// rpcDeletePrinterResponse carries the [out] parameters and return value of RpcDeletePrinter.
type rpcDeletePrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrinter calls RpcDeletePrinter (opnum 6) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrinter(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE) (err error) {
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
