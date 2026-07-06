package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncWritePrinterRequest carries the [in] parameters of RpcAsyncWritePrinter.
type rpcAsyncWritePrinterRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	PBuf     []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcAsyncWritePrinterRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncWritePrinter }

// rpcAsyncWritePrinterResponse carries the [out] parameters and return value of RpcAsyncWritePrinter.
type rpcAsyncWritePrinterResponse struct {
	PcWritten ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncWritePrinter calls RpcAsyncWritePrinter (opnum 12) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncWritePrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pBuf []uint8, cbBuf ndr.DWORD) (PcWritten ndr.DWORD, err error) {
	req := &rpcAsyncWritePrinterRequest{
		HPrinter: hPrinter,
		PBuf:     pBuf,
		CbBuf:    cbBuf,
	}
	var resp rpcAsyncWritePrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncWritePrinter: %w", err)
		return
	}
	PcWritten = resp.PcWritten
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncWritePrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
