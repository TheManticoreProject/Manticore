package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncSetFormRequest carries the [in] parameters of RpcAsyncSetForm.
type rpcAsyncSetFormRequest struct {
	HPrinter           mspar.PRINTER_HANDLE
	PFormName          ndr.WSTR
	PFormInfoContainer mspar.FORM_CONTAINER
}

func (*rpcAsyncSetFormRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncSetForm }

// rpcAsyncSetFormResponse carries the [out] parameters and return value of RpcAsyncSetForm.
type rpcAsyncSetFormResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncSetForm calls RpcAsyncSetForm (opnum 24) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncSetForm(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pFormName ndr.WSTR, pFormInfoContainer mspar.FORM_CONTAINER) (err error) {
	req := &rpcAsyncSetFormRequest{
		HPrinter:           hPrinter,
		PFormName:          pFormName,
		PFormInfoContainer: pFormInfoContainer,
	}
	var resp rpcAsyncSetFormResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncSetForm: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncSetForm failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
