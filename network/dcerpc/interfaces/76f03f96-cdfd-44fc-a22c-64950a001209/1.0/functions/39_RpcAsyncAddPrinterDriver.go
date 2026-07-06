package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncAddPrinterDriverRequest carries the [in] parameters of RpcAsyncAddPrinterDriver.
type rpcAsyncAddPrinterDriverRequest struct {
	PName            *ndr.WSTR `ndr:"unique"`
	PDriverContainer mspar.DRIVER_CONTAINER
	DwFileCopyFlags  ndr.DWORD
}

func (*rpcAsyncAddPrinterDriverRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncAddPrinterDriver
}

// rpcAsyncAddPrinterDriverResponse carries the [out] parameters and return value of RpcAsyncAddPrinterDriver.
type rpcAsyncAddPrinterDriverResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncAddPrinterDriver calls RpcAsyncAddPrinterDriver (opnum 39) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncAddPrinterDriver(rpc ndr.Invoker, pName *ndr.WSTR, pDriverContainer mspar.DRIVER_CONTAINER, dwFileCopyFlags ndr.DWORD) (err error) {
	req := &rpcAsyncAddPrinterDriverRequest{
		PName:            pName,
		PDriverContainer: pDriverContainer,
		DwFileCopyFlags:  dwFileCopyFlags,
	}
	var resp rpcAsyncAddPrinterDriverResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncAddPrinterDriver: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncAddPrinterDriver failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
