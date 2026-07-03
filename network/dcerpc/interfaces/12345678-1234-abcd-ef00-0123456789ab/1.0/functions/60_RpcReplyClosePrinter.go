package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcReplyClosePrinterRequest carries the [in] parameters of RpcReplyClosePrinter.
type rpcReplyClosePrinterRequest struct {
	PhNotify msrprn.PRINTER_HANDLE
}

func (*rpcReplyClosePrinterRequest) Opnum() uint16 { return winspool.OpnumRpcReplyClosePrinter }

// rpcReplyClosePrinterResponse carries the [out] parameters and return value of RpcReplyClosePrinter.
type rpcReplyClosePrinterResponse struct {
	PhNotify msrprn.PRINTER_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// RpcReplyClosePrinter calls RpcReplyClosePrinter (opnum 60) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcReplyClosePrinter(rpc ndr.Invoker, phNotify msrprn.PRINTER_HANDLE) (PhNotify msrprn.PRINTER_HANDLE, err error) {
	req := &rpcReplyClosePrinterRequest{
		PhNotify: phNotify,
	}
	var resp rpcReplyClosePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcReplyClosePrinter: %w", err)
		return
	}
	PhNotify = resp.PhNotify
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcReplyClosePrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
