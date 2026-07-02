package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEndDocPrinterRequest carries the [in] parameters of RpcEndDocPrinter.
type rpcEndDocPrinterRequest struct {
	HPrinter structures.PRINTER_HANDLE
}

func (*rpcEndDocPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcEndDocPrinter }

// rpcEndDocPrinterResponse carries the [out] parameters and return value of RpcEndDocPrinter.
type rpcEndDocPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcEndDocPrinter calls RpcEndDocPrinter (opnum 23) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEndDocPrinter(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE) (err error) {
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
