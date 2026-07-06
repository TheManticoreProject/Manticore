package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncStartDocPrinterRequest carries the [in] parameters of RpcAsyncStartDocPrinter.
type rpcAsyncStartDocPrinterRequest struct {
	HPrinter          mspar.PRINTER_HANDLE
	PDocInfoContainer mspar.DOC_INFO_CONTAINER
}

func (*rpcAsyncStartDocPrinterRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncStartDocPrinter
}

// rpcAsyncStartDocPrinterResponse carries the [out] parameters and return value of RpcAsyncStartDocPrinter.
type rpcAsyncStartDocPrinterResponse struct {
	PJobId ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncStartDocPrinter calls RpcAsyncStartDocPrinter (opnum 10) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncStartDocPrinter(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pDocInfoContainer mspar.DOC_INFO_CONTAINER) (PJobId ndr.DWORD, err error) {
	req := &rpcAsyncStartDocPrinterRequest{
		HPrinter:          hPrinter,
		PDocInfoContainer: pDocInfoContainer,
	}
	var resp rpcAsyncStartDocPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncStartDocPrinter: %w", err)
		return
	}
	PJobId = resp.PJobId
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncStartDocPrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
