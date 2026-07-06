package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncReadPrinterRequest carries the [in] parameters of RpcAsyncReadPrinter.
type rpcAsyncReadPrinterRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	CbBuf    ndr.DWORD
}

func (*rpcAsyncReadPrinterRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncReadPrinter }

// rpcAsyncReadPrinterResponse carries the [out] parameters and return value of RpcAsyncReadPrinter.
type rpcAsyncReadPrinterResponse struct {
	PBuf          []uint8 `ndr:"ref,size_is=CbBuf"`
	PcNoBytesRead ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcAsyncReadPrinter calls RpcAsyncReadPrinter (opnum 68) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncReadPrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, cbBuf ndr.DWORD) (PBuf []uint8, PcNoBytesRead ndr.DWORD, err error) {
	req := &rpcAsyncReadPrinterRequest{
		HPrinter: hPrinter,
		CbBuf:    cbBuf,
	}
	var resp rpcAsyncReadPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncReadPrinter: %w", err)
		return
	}
	PBuf = resp.PBuf
	PcNoBytesRead = resp.PcNoBytesRead
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncReadPrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
