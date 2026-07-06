package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncDeletePrinterICRequest carries the [in] parameters of RpcAsyncDeletePrinterIC.
type rpcAsyncDeletePrinterICRequest struct {
	PhPrinterIC mspar.GDI_HANDLE
}

func (*rpcAsyncDeletePrinterICRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncDeletePrinterIC
}

// rpcAsyncDeletePrinterICResponse carries the [out] parameters and return value of RpcAsyncDeletePrinterIC.
type rpcAsyncDeletePrinterICResponse struct {
	PhPrinterIC mspar.GDI_HANDLE
	Status      ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeletePrinterIC calls RpcAsyncDeletePrinterIC (opnum 37) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeletePrinterIC(rpc ndr.Invoker, phPrinterIC mspar.GDI_HANDLE) (PhPrinterIC mspar.GDI_HANDLE, err error) {
	req := &rpcAsyncDeletePrinterICRequest{
		PhPrinterIC: phPrinterIC,
	}
	var resp rpcAsyncDeletePrinterICResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeletePrinterIC: %w", err)
		return
	}
	PhPrinterIC = resp.PhPrinterIC
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeletePrinterIC failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
