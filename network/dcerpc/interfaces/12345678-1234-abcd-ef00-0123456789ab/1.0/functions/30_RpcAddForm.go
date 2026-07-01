package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAddFormRequest carries the [in] parameters of RpcAddForm.
type rpcAddFormRequest struct {
	HPrinter           structures.PRINTER_HANDLE
	PFormInfoContainer structures.FORM_CONTAINER
}

func (*rpcAddFormRequest) Opnum() uint16 { return winspool.OpnumRpcAddForm }

// rpcAddFormResponse carries the [out] parameters and return value of RpcAddForm.
type rpcAddFormResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddForm calls RpcAddForm (opnum 30) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddForm(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pFormInfoContainer structures.FORM_CONTAINER) (err error) {
	req := &rpcAddFormRequest{
		HPrinter:           hPrinter,
		PFormInfoContainer: pFormInfoContainer,
	}
	var resp rpcAddFormResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddForm: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddForm failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
