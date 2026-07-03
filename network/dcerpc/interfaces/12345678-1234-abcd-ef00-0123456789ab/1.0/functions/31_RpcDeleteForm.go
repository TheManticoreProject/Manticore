package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcDeleteFormRequest carries the [in] parameters of RpcDeleteForm.
type rpcDeleteFormRequest struct {
	HPrinter  msrprn.PRINTER_HANDLE
	PFormName ndr.WSTR
}

func (*rpcDeleteFormRequest) Opnum() uint16 { return winspool.OpnumRpcDeleteForm }

// rpcDeleteFormResponse carries the [out] parameters and return value of RpcDeleteForm.
type rpcDeleteFormResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeleteForm calls RpcDeleteForm (opnum 31) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeleteForm(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pFormName ndr.WSTR) (err error) {
	req := &rpcDeleteFormRequest{
		HPrinter:  hPrinter,
		PFormName: pFormName,
	}
	var resp rpcDeleteFormResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeleteForm: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeleteForm failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
