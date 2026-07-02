package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcSetFormRequest carries the [in] parameters of RpcSetForm.
type rpcSetFormRequest struct {
	HPrinter           structures.PRINTER_HANDLE
	PFormName          ndr.WSTR
	PFormInfoContainer structures.FORM_CONTAINER
}

func (*rpcSetFormRequest) Opnum() uint16 { return winspool.OpnumRpcSetForm }

// rpcSetFormResponse carries the [out] parameters and return value of RpcSetForm.
type rpcSetFormResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcSetForm calls RpcSetForm (opnum 33) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSetForm(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pFormName ndr.WSTR, pFormInfoContainer structures.FORM_CONTAINER) (err error) {
	req := &rpcSetFormRequest{
		HPrinter:           hPrinter,
		PFormName:          pFormName,
		PFormInfoContainer: pFormInfoContainer,
	}
	var resp rpcSetFormResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSetForm: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSetForm failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
