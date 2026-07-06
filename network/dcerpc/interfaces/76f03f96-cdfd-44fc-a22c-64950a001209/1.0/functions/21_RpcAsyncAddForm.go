package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncAddFormRequest carries the [in] parameters of RpcAsyncAddForm.
type rpcAsyncAddFormRequest struct {
	HPrinter           mspar.PRINTER_HANDLE
	PFormInfoContainer mspar.FORM_CONTAINER
}

func (*rpcAsyncAddFormRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncAddForm }

// rpcAsyncAddFormResponse carries the [out] parameters and return value of RpcAsyncAddForm.
type rpcAsyncAddFormResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncAddForm calls RpcAsyncAddForm (opnum 21) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncAddForm(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pFormInfoContainer mspar.FORM_CONTAINER) (err error) {
	req := &rpcAsyncAddFormRequest{
		HPrinter:           hPrinter,
		PFormInfoContainer: pFormInfoContainer,
	}
	var resp rpcAsyncAddFormResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncAddForm: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncAddForm failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
