package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcWaitForPrinterChangeRequest carries the [in] parameters of RpcWaitForPrinterChange.
type rpcWaitForPrinterChangeRequest struct {
	HPrinter structures.PRINTER_HANDLE
	Flags    ndr.DWORD
}

func (*rpcWaitForPrinterChangeRequest) Opnum() uint16 { return winspool.OpnumRpcWaitForPrinterChange }

// rpcWaitForPrinterChangeResponse carries the [out] parameters and return value of RpcWaitForPrinterChange.
type rpcWaitForPrinterChangeResponse struct {
	PFlags ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// RpcWaitForPrinterChange calls RpcWaitForPrinterChange (opnum 28) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcWaitForPrinterChange(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, flags ndr.DWORD) (PFlags ndr.DWORD, err error) {
	req := &rpcWaitForPrinterChangeRequest{
		HPrinter: hPrinter,
		Flags:    flags,
	}
	var resp rpcWaitForPrinterChangeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWaitForPrinterChange: %w", err)
		return
	}
	PFlags = resp.PFlags
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcWaitForPrinterChange failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
