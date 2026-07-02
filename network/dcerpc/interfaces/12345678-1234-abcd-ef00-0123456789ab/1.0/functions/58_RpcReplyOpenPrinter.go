package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcReplyOpenPrinterRequest carries the [in] parameters of RpcReplyOpenPrinter.
type rpcReplyOpenPrinterRequest struct {
	PMachine        ndr.WSTR
	DwPrinterRemote ndr.DWORD
	DwType          ndr.DWORD
	CbBuffer        ndr.DWORD
	PBuffer         []uint8 `ndr:"unique,size_is=CbBuffer"`
}

func (*rpcReplyOpenPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcReplyOpenPrinter }

// rpcReplyOpenPrinterResponse carries the [out] parameters and return value of RpcReplyOpenPrinter.
type rpcReplyOpenPrinterResponse struct {
	PhPrinterNotify structures.PRINTER_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcReplyOpenPrinter calls RpcReplyOpenPrinter (opnum 58) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcReplyOpenPrinter(rpc ndr.Invoker, pMachine ndr.WSTR, dwPrinterRemote ndr.DWORD, dwType ndr.DWORD, cbBuffer ndr.DWORD, pBuffer []uint8) (PhPrinterNotify structures.PRINTER_HANDLE, err error) {
	req := &rpcReplyOpenPrinterRequest{
		PMachine:        pMachine,
		DwPrinterRemote: dwPrinterRemote,
		DwType:          dwType,
		CbBuffer:        cbBuffer,
		PBuffer:         pBuffer,
	}
	var resp rpcReplyOpenPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcReplyOpenPrinter: %w", err)
		return
	}
	PhPrinterNotify = resp.PhPrinterNotify
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcReplyOpenPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
