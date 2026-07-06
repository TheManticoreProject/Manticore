package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncDeleteFormRequest carries the [in] parameters of RpcAsyncDeleteForm.
type rpcAsyncDeleteFormRequest struct {
	HPrinter  mspar.PRINTER_HANDLE
	PFormName ndr.WSTR
}

func (*rpcAsyncDeleteFormRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncDeleteForm }

// rpcAsyncDeleteFormResponse carries the [out] parameters and return value of RpcAsyncDeleteForm.
type rpcAsyncDeleteFormResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncDeleteForm calls RpcAsyncDeleteForm (opnum 22) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncDeleteForm(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pFormName ndr.WSTR) (err error) {
	req := &rpcAsyncDeleteFormRequest{
		HPrinter:  hPrinter,
		PFormName: pFormName,
	}
	var resp rpcAsyncDeleteFormResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncDeleteForm: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncDeleteForm failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
