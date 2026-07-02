package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcRouterReplyPrinterRequest carries the [in] parameters of RpcRouterReplyPrinter.
type rpcRouterReplyPrinterRequest struct {
	HNotify  structures.PRINTER_HANDLE
	FdwFlags ndr.DWORD
	CbBuffer ndr.DWORD
	PBuffer  []uint8 `ndr:"unique,size_is=CbBuffer"`
}

func (*rpcRouterReplyPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcRouterReplyPrinter }

// rpcRouterReplyPrinterResponse carries the [out] parameters and return value of RpcRouterReplyPrinter.
type rpcRouterReplyPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcRouterReplyPrinter calls RpcRouterReplyPrinter (opnum 59) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcRouterReplyPrinter(rpc ndr.Invoker, hNotify structures.PRINTER_HANDLE, fdwFlags ndr.DWORD, cbBuffer ndr.DWORD, pBuffer []uint8) (err error) {
	req := &rpcRouterReplyPrinterRequest{
		HNotify:  hNotify,
		FdwFlags: fdwFlags,
		CbBuffer: cbBuffer,
		PBuffer:  pBuffer,
	}
	var resp rpcRouterReplyPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcRouterReplyPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcRouterReplyPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
